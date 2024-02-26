package convert

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"go.expect.digital/mf2"
	ast "go.expect.digital/mf2/parse"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/po"
	"golang.org/x/text/language"
)

// ---------------------------------------PO->Translation---------------------------------------

// FromPo converts a byte slice representing a PO file to a model.Translation structure.
// TODO: We need to find a way to preserve Plural-Forms from PO file.
func FromPo(b []byte, originalOverride *bool) (model.Translation, error) {
	file, err := po.Parse(b)
	if err != nil {
		return model.Translation{}, fmt.Errorf("parse portableObject file: %w", err)
	}

	var lang language.Tag

	if langStr := file.Headers.Get("Language"); langStr != "" {
		lang, err = language.Parse(langStr)
		if err != nil {
			return model.Translation{}, fmt.Errorf("parse language tag: %w", err)
		}
	}

	translation := model.Translation{
		Language: lang,
		Messages: make([]model.Message, 0, len(file.Messages)),
		Original: isOriginalPO(file, originalOverride),
	}

	var (
		getStatus   func(po.Message) model.MessageStatus // status getter based on originality
		getMessages func(po.Message) []string            // messages getter based on originality
	)

	switch translation.Original {
	// on original, all messages are considered translated
	// messages are get from msgid and msgid_plural if exists.
	case true:
		getStatus = func(po.Message) model.MessageStatus { return model.MessageStatusTranslated }
		getMessages = func(n po.Message) []string {
			if n.MsgIDPlural != "" {
				return []string{n.MsgID, n.MsgIDPlural}
			}

			return []string{n.MsgID}
		}
	// on non original, all messages are considered untranslated or fuzzy
	// messages are get from msgstr and msgstr[*] if exists.
	case false:
		getStatus = func(n po.Message) model.MessageStatus {
			if slices.Contains(n.Flags, "fuzzy") {
				return model.MessageStatusFuzzy
			}

			return model.MessageStatusUntranslated
		}
		getMessages = func(n po.Message) []string { return n.MsgStr }
	}

	for _, node := range file.Messages {
		mf2Msg, err := msgNodeToMF2(node, getMessages)
		if err != nil {
			return model.Translation{}, fmt.Errorf("convert message node to mf2 format: %w", err)
		}

		translation.Messages = append(translation.Messages, model.Message{
			ID:          node.MsgID,
			PluralID:    node.MsgIDPlural,
			Description: strings.Join(node.ExtractedComments, "\n"),
			Positions:   node.References,
			Message:     mf2Msg,
			Status:      getStatus(node),
		})
	}

	return translation, nil
}

// isOriginalPO function determines whether a PO file is an original or a translation.
func isOriginalPO(portableObject po.PO, override *bool) bool {
	// If override is not nil, use it.
	if override != nil {
		return *override
	}

	// NOTE: Based on my research, all original PO files have empty language.
	// So that could be a way to determine originality.
	// Further research is needed to confirm this.
	if portableObject.Headers.Get("Language") == "" {
		return true
	}

	// When dealing with original PO files, all messages are always empty.
	allEmpty := func(msgs []po.Message) bool {
		for _, node := range msgs {
			for _, msg := range node.MsgStr {
				if msg != "" {
					return false
				}
			}
		}

		return true
	}

	// NOTE: originality can also be determined by file extension.
	// .pot == original and .po == translation,
	// but we don't preserve file extension, so we can't use this method for now.

	return allEmpty(portableObject.Messages)
}

// TODO: Can every convert use that ?
var placeholderFormats = map[string]*regexp.Regexp{
	"pythonVar":    regexp.MustCompile(`%\((\w+)\)([sd])`), // hello %(var)s | hello %(var)d
	"printf":       regexp.MustCompile(`%(s|d|f)`),         // hello %s | hello %d | hello %f
	"bracketVar":   regexp.MustCompile(`\{(\w+)\}`),        // hello {var} | hello {0}
	"emptyBracket": regexp.MustCompile(`\{\}`),             // hello {}
}

// msgNodeToMF2 function converts a po.MessageNode to a MessageFormat2 string.
func msgNodeToMF2(node po.Message, getMessages func(po.Message) []string) (string, error) {
	mfBuilder := mf2.NewBuilder()
	placeholders := make(map[string]struct{}) // map of placeholders to avoid duplicates, only for plural messages

	// default build function, for message without placeholders
	build := func(mfBuilder *mf2.Builder, msg string, _ map[string]struct{}) { mfBuilder.Text(msg) }

	// look for format flag, e.g. #, python-format, c-format, no-python-format, etc.
	formatFlagIdx := slices.IndexFunc(node.Flags, func(flag string) bool {
		return strings.HasSuffix(flag, "-format")
	})

	if formatFlagIdx != -1 {
		mfBuilder.Local("format", mf2.Literal(node.Flags[formatFlagIdx])) // capture format flag
	}

	switch messages := getMessages(node); len(messages) {
	case 1: // singular message
		if formatFlagIdx != -1 && !strings.HasPrefix(node.Flags[formatFlagIdx], "no-") { // with placeholders
			build = textWithPlaceholders
		}

		build(mfBuilder, messages[0], placeholders)

	default: // plural message
		mfBuilder.Match(mf2.Var("count")) // match to arbitrary variable name

		if formatFlagIdx != -1 && !strings.HasPrefix(node.Flags[formatFlagIdx], "no-") { // with placeholders
			build = textWithPlaceholders
		}

		for i := range messages {
			if i == len(messages)-1 {
				mfBuilder.Keys("*")
			} else {
				mfBuilder.Keys(i + 1) // arbitrary key names to match to //TODO can be derived from plural forms
			}

			build(mfBuilder, messages[i], placeholders)
		}
	}

	mf2String, err := mfBuilder.Build()
	if err != nil {
		return "", fmt.Errorf("build mf2 string: %w", err)
	}

	return mf2String, nil
}

// textWithPlaceholders processes a message string, identifies placeholders, and adds them to a Builder.
func textWithPlaceholders(mfBuilder *mf2.Builder, msg string, placeholders map[string]struct{}) {
	for name, re := range placeholderFormats {
		var currentIdx int

		allIndices := re.FindAllStringIndex(msg, -1)
		for i, indices := range allIndices {
			// Add text before the first placeholder, if any.
			text := msg[currentIdx:indices[0]]
			if text != "" {
				mfBuilder.Text(text)
			}

			originalVariable := msg[indices[0]:indices[1]]

			var mf2Variable string

			switch name {
			case "printf", "emptyBracket":
				mf2Variable = fmt.Sprintf("ph%d", i) // %d|%s|%f -> ph0|ph1|ph2...
			case "pythonVar":
				mf2Variable = originalVariable[2 : len(originalVariable)-2] // %(var)s -> var
			case "bracketVar":
				mf2Variable = originalVariable[1 : len(originalVariable)-1] // {var} -> var
			}

			// Avoid adding duplicate locals variables when working with plural messages
			if _, ok := placeholders[mf2Variable]; !ok {
				mfBuilder.Local(mf2Variable, mf2.Literal(originalVariable))

				placeholders[mf2Variable] = struct{}{}
			}

			mfBuilder.Expr(mf2.Var(mf2Variable))

			currentIdx = indices[1]

			// If this is the last placeholder, add the remaining text after it, if any.
			if i == len(allIndices)-1 {
				text := msg[currentIdx:]
				if text != "" {
					mfBuilder.Text(text)
				}

				return
			}
		}
	}

	// placeholder flag is present, but no placeholders found
	mfBuilder.Text(msg)
}

// ---------------------------------------Translation->PO---------------------------------------

// ToPo converts a model.Translation structure to a byte slice representing a PO file.
func ToPo(t model.Translation) ([]byte, error) {
	portableObject := po.PO{Messages: make([]po.Message, 0, len(t.Messages))}

	if !t.Original {
		portableObject.Headers = append(portableObject.Headers, po.Header{Name: "Language", Value: t.Language.String()})
	}

	var placeholders map[ast.Variable]string // MF2Variable:OriginalVariable, only for complex messages

	// patternsToMsg function converts a slice of patterns to a PO msgstr.
	patternsToMsg := func(patterns []ast.Pattern) string {
		var text string

		for _, p := range patterns {
			switch p := p.(type) {
			case ast.TextPattern:
				text += string(p)
			case ast.Expression:
				text += placeholders[p.Operand.(ast.Variable)] //nolint:forcetypeassert // operand is always of type Variable
			}
		}

		return text
	}

	// If original, msgstr are empty
	if t.Original {
		patternsToMsg = func([]ast.Pattern) string { return "" }
	}

	unquoteLiteral := func(l ast.Literal) string { return strings.ReplaceAll(l.String(), "|", "") }

	for _, message := range t.Messages {
		// Build po.MessageNode, from model.Message.
		poMsg := po.Message{
			MsgID:       message.ID,
			MsgIDPlural: message.PluralID,
			References:  message.Positions,
			MsgStr:      make([]string, 0, 1), // At least one string will always be present.
		}

		if message.Description != "" {
			poMsg.ExtractedComments = strings.Split(message.Description, "\n")
		}

		if message.Status == model.MessageStatusFuzzy {
			poMsg.Flags = append(poMsg.Flags, "fuzzy")
		}

		// Parse mf2 message.

		tree, err := ast.Parse(message.Message)
		if err != nil {
			return nil, fmt.Errorf("parse mf2 message: %w", err)
		}

		switch message := tree.Message.(type) {
		case nil:
			poMsg.MsgStr = append(poMsg.MsgStr, "") // no mf2 message
		case ast.SimpleMessage:
			poMsg.MsgStr = append(poMsg.MsgStr, patternsToMsg(message))
		case ast.ComplexMessage:
			// Declarations
			placeholders = make(map[ast.Variable]string, len(message.Declarations))

			for _, decl := range message.Declarations {
				decl := decl.(ast.LocalDeclaration) //nolint:forcetypeassert // no other types of declarations are used in convert

				//nolint:forcetypeassert // All declarations are of type LiteralExpression.
				switch decl.Variable.String() {
				case "$format": // flag
					poMsg.Flags = append(poMsg.Flags, unquoteLiteral(decl.Expression.Operand.(ast.Literal)))
				default: // placeholder
					placeholders[decl.Variable] = unquoteLiteral(decl.Expression.Operand.(ast.Literal))
				}
			}

			// Body
			switch body := message.ComplexBody.(type) {
			case ast.Matcher:
				poMsg.MsgStr = make([]string, 0, len(body.Variants))
				for _, variant := range body.Variants {
					poMsg.MsgStr = append(poMsg.MsgStr, patternsToMsg(variant.QuotedPattern))
				}
			case ast.QuotedPattern:
				poMsg.MsgStr = append(poMsg.MsgStr, patternsToMsg(body))
			}
		}

		portableObject.Messages = append(portableObject.Messages, poMsg)
	}

	return portableObject.Marshal(), nil
}
