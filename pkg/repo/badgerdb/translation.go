package badgerdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
	"golang.org/x/text/language"
)

const translationPrefix = "translation:"

// translationKey converts a serviceID and language to a BadgerDB key with prefix.
func translationKey(serviceID uuid.UUID, language language.Tag) []byte {
	return []byte(fmt.Sprintf("%s%s:%s", translationPrefix, serviceID, language))
}

// SaveTranslation handles both Create and Update.
func (r *Repo) SaveTranslation(ctx context.Context, serviceID uuid.UUID, translation *model.Translation) error {
	if r.tx != nil { // use existing tx
		return r.saveTranslation(ctx, serviceID, translation)
	}

	return r.Tx(ctx, func(ctx context.Context, rp repo.Repo) error { // create new tx
		return rp.SaveTranslation(ctx, serviceID, translation) //nolint:wrapcheck
	})
}

func (r *Repo) saveTranslation(ctx context.Context, serviceID uuid.UUID, translation *model.Translation) error {
	_, err := r.LoadService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("repo: load service: %w", err)
	}

	b, err := json.Marshal(translation)
	if err != nil {
		return fmt.Errorf("marshal translation: %w", err)
	}

	if err := r.tx.Set(translationKey(serviceID, translation.Language), b); err != nil {
		return fmt.Errorf("repo: set translation: %w", err)
	}

	return nil
}

// LoadTranslations retrieves translations from db based on serviceID and LoadMessageOpts.
func (r *Repo) LoadTranslations(ctx context.Context, serviceID uuid.UUID, opts repo.LoadTranslationsOpts,
) (model.Translations, error) {
	if _, err := r.LoadService(ctx, serviceID); errors.Is(err, repo.ErrNotFound) {
		return nil, nil // Empty translation.messages for this service (Not an error)
	} else if err != nil {
		return nil, fmt.Errorf("repo: load service: %w", err)
	}

	// load all translations if languages are not provided.
	if len(opts.FilterLanguages) == 0 {
		translations, err := r.loadTranslations(serviceID)
		if err != nil {
			return nil, fmt.Errorf("load translations by service '%s': %w", serviceID, err)
		}

		return translations, nil
	}

	// load translations based on provided languages.
	translations, err := r.loadTranslationsByLang(serviceID, opts.FilterLanguages)
	if err != nil {
		return nil, fmt.Errorf("load translations by languages: %w", err)
	}

	return translations, nil
}

// loadTranslationsByLang returns translations for service based on provided languages.
func (r *Repo) loadTranslationsByLang(serviceID uuid.UUID, languages []language.Tag,
) (model.Translations, error) {
	translations := make([]model.Translation, 0, len(languages))

	if err := r.db.View(func(txn *badger.Txn) error {
		for _, lang := range languages {
			var translation model.Translation

			item, txErr := txn.Get(translationKey(serviceID, lang))
			switch {
			default:
				if valErr := getValue(item, &translation); valErr != nil {
					return fmt.Errorf("get translation by language '%s': %w", lang, valErr)
				}

				translations = append(translations, translation)
			case errors.Is(txErr, badger.ErrKeyNotFound):
				return nil // No Translations for this language (Not an error)
			case txErr != nil:
				return fmt.Errorf("transaction: get translations by language '%s': %w", lang, txErr)
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return translations, nil
}

// loadTranslations returns all translations for service.
func (r *Repo) loadTranslations(serviceID uuid.UUID) (model.Translations, error) {
	keyPrefix := []byte(translationPrefix + serviceID.String())

	var translations []model.Translation

	if err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(keyPrefix); it.ValidForPrefix(keyPrefix); it.Next() {
			translation := model.Translation{}

			if err := getValue(it.Item(), &translation); err != nil {
				return fmt.Errorf("transaction: get value: %w", err)
			}

			translations = append(translations, translation)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("repo: db view: %w", err)
	}

	return translations, nil
}
