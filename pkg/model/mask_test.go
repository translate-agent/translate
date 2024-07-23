package model

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"go.expect.digital/translate/pkg/testutil/expect"

	"golang.org/x/text/language"
)

// deepCopy makes a deep copy of src and returns it.
func deepCopy[T any](t *testing.T, src T) (dst T) { //nolint:ireturn
	t.Helper()

	data, err := json.Marshal(src)
	if err != nil {
		t.Error(err)
		return
	}

	expect.NoError(t, json.Unmarshal(data, &dst))

	return
}

//nolint:gocognit
func Test_UpdateNestedStructFromMask(t *testing.T) {
	t.Parallel()

	//nolint: govet
	type nestedStruct struct {
		A int
		B string
		C struct {
			D float32
			E string
			F struct {
				G []string
			}
			H []struct {
				I string
			}
		}
		J struct {
			K map[string]string
		}
		L *string
	}

	// Generate random source and destination structs
	var src, dst nestedStruct

	expect.NoError(t, gofakeit.Struct(&src))
	expect.NoError(t, gofakeit.Struct(&dst))

	tests := []struct {
		assertFunc func(t *testing.T, src, dst, original nestedStruct)
		name       string
		mask       Mask
	}{
		{
			// Update one top-level field
			name: "Update A int",
			mask: []string{"A"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				// Check if field is updated
				expect.Equal(t, src.A, dst.A)

				// Reset field to original value, and perform full check, to ensure that nothing else was changed
				dst.A = original.A

				if !reflect.DeepEqual(original, dst) {
					t.Errorf("want %v, got %v", original, dst)
				}
			},
		},
		{
			// Update two top-level fields
			name: "Update A and B int and string",
			mask: []string{"A", "B"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				expect.Equal(t, src.A, dst.A)

				dst.A, dst.B = original.A, original.B

				if !reflect.DeepEqual(original, dst) {
					t.Errorf("\nwant %v\ngot  %v", original, dst)
				}
			},
		},
		{
			// Update whole top-level struct
			name: "Update C struct",
			mask: []string{"C"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				if !reflect.DeepEqual(src.C, dst.C) {
					t.Errorf("want %v, got %v", src.C, dst.C)
					return
				}

				dst.C = original.C

				if !reflect.DeepEqual(original, dst) {
					t.Errorf("\nwant %v\ngot  %v", original, dst)
				}
			},
		},
		{
			// Update top-level field of a nested struct
			name: "Update C.D struct.float",
			mask: []string{"C.D"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				if dst.C.D-src.C.D >= 0.01 {
					t.Errorf("want %f, got %f", src.C.D, dst.C.D)
				}

				dst.C.D = original.C.D

				if !reflect.DeepEqual(original, dst) {
					t.Errorf("\nwant %v\ngot  %v", original, dst)
				}
			},
		},
		{
			// Update a whole nested struct
			name: "Update C.F struct.struct",
			mask: []string{"C.F"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				if !reflect.DeepEqual(src.C.F, dst.C.F) {
					t.Errorf("want %v, got %v", src.C.F, dst.C.F)
					return
				}

				dst.C.F = original.C.F

				if !reflect.DeepEqual(original, dst) {
					t.Errorf("\nwant %v\ngot  %v", original, dst)
				}
			},
		},
		{
			// Update slice of strings in a double nested struct (merge two slices)
			name: "Update C.F.G struct.struct.[]string",
			mask: []string{"C.F.G"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				// Check if all elements from src and dst are in result
				for _, v := range src.C.F.G {
					if !slices.Contains(dst.C.F.G, v) {
						t.Errorf("want %v to contain %s", dst.C.F.G, v)
					}
				}

				dst.C.F.G = original.C.F.G

				if !reflect.DeepEqual(original, dst) {
					t.Errorf("\nwant %v\ngot  %v", original, dst)
				}
			},
		},
		{
			// Update slice of custom structs in a nested struct. (Merge two slices)
			name: "Update C.H struct.struct.[]struct",
			mask: []string{"C.H"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				// Check if all elements from src and dst are in result
				for _, srcElem := range src.C.H {
					if !slices.Contains(dst.C.H, srcElem) {
						t.Errorf("want %v to contain %v", dst.C.H, srcElem)
						return
					}
				}

				dst.C.H = original.C.H

				if !reflect.DeepEqual(original, dst) {
					t.Errorf("\nwant %v\ngot  %v", original, dst)
				}
			},
		},
		{
			// Update map of strings in a nested field of struct. (Merge two maps)
			name: "Update J.K struct.map[string]string",
			mask: []string{"J.K"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				// Check if all keys from src and dst are in result
				for srcKey := range src.J.K {
					if _, ok := dst.J.K[srcKey]; !ok {
						t.Errorf("want %v to contain %s", dst.J.K, srcKey)
						return
					}
				}

				dst.J.K = original.J.K

				if !reflect.DeepEqual(original, dst) {
					t.Errorf("\nwant %v\ngot  %v", original, dst)
				}
			},
		},
		{
			// Update top level pointer to string
			name: "Update L *string",
			mask: []string{"L"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				expect.Equal(t, src.L, dst.L)

				dst.L = original.L

				if !reflect.DeepEqual(original, dst) {
					t.Errorf("\nwant %v\ngot  %v", original, dst)
				}
			},
		},
		{
			// Update all fields
			name: "Update All",
			mask: nil,
			assertFunc: func(t *testing.T, src, dst, _ nestedStruct) {
				if !reflect.DeepEqual(src, dst) {
					t.Errorf("\nwant %v\ngot  %v", src, dst)
				}
			},
		},
		{
			// No Paths in FieldMask. Updates nothing.
			name: "Update Nothing Empty Paths",
			mask: Mask{},
			assertFunc: func(t *testing.T, _, dst, original nestedStruct) {
				if !reflect.DeepEqual(original, dst) {
					t.Errorf("want %v, got %v", original, dst)
				}
			},
		},
		{
			// Random path in FieldMask. Updates nothing.
			name: "Update Nothing Random Path",
			mask: Mask{"random_path"},
			assertFunc: func(t *testing.T, _, dst, original nestedStruct) {
				if !reflect.DeepEqual(original, dst) {
					t.Errorf("want %v, got %v", original, dst)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Make deep copies of structs for deterministic tests
			original := deepCopy(t, dst)
			dstCopy, srcCopy := deepCopy(t, dst), deepCopy(t, src)

			Update(&srcCopy, &dstCopy, test.mask)
			test.assertFunc(t, srcCopy, dstCopy, original)
		})
	}
}

func Test_UpdateServiceFromMask(t *testing.T) {
	t.Parallel()

	// Generate random source and destination structs
	var srcService, dstService Service

	expect.NoError(t, gofakeit.Struct(&srcService))
	expect.NoError(t, gofakeit.Struct(&dstService))

	tests := []struct {
		assertFunc func(t *testing.T, srcService, dstService, original Service)
		name       string
		wantErr    error
		fieldMask  Mask
	}{
		// positive tests
		{
			name:      "Update Name",
			fieldMask: Mask{"Name"},
			assertFunc: func(t *testing.T, srcService, dstService, original Service) {
				// Same ID updated name
				if original.ID != dstService.ID {
					t.Errorf("want id '%s', got '%s'", original.ID, dstService.ID)
				}

				if srcService.Name != dstService.Name {
					t.Errorf("want name '%s', got '%s'", srcService.Name, dstService.Name)
				}
			},
		},
		{
			name:      "Update All",
			fieldMask: nil,
			assertFunc: func(t *testing.T, srcService, dstService, original Service) {
				// Same ID updated name, as ID cannot be updated, and service has only two fields.
				if original.ID != dstService.ID {
					t.Errorf("want %s, got %s", original.ID, dstService.ID)
				}

				if srcService.Name != dstService.Name {
					t.Errorf("want %s, got %s", srcService.Name, dstService.Name)
				}
			},
		},
		{
			name:      "Nothing to Update Empty Paths",
			fieldMask: Mask{},
			assertFunc: func(t *testing.T, _, dstService, original Service) {
				// Same ID and name, as nothing was updated
				if !reflect.DeepEqual(dstService, original) {
					t.Errorf("want %v, got %v", dstService, original)
				}
			},
		},
		{
			name:      "Nothing to Update Random Path",
			fieldMask: Mask{"random_path"},
			assertFunc: func(t *testing.T, _, dstService, original Service) {
				// Same ID and name, as nothing was updated
				if !reflect.DeepEqual(dstService, original) {
					t.Errorf("want %v, got %v", dstService, original)
				}
			},
		},
		// negative tests
		{
			name:      "Try to update ID",
			fieldMask: Mask{"ID"},
			wantErr:   errors.New("\"id\" is not allowed in field mask"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Make deep copies of structs for deterministic tests
			original := deepCopy(t, dstService)
			dstCopy, srcCopy := deepCopy(t, dstService), deepCopy(t, srcService)

			err := UpdateService(&srcCopy, &dstCopy, test.fieldMask)

			if test.wantErr != nil {
				if err == nil || err.Error() != test.wantErr.Error() {
					t.Errorf("want %s, got %s", test.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			test.assertFunc(t, srcCopy, dstCopy, original)
		})
	}
}

func Test_UpdateTranslationFromMask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fieldMask      Mask
		srcTranslation Translation
		dstTranslation Translation
		want           Translation
	}{
		{
			name:           "Update Original",
			fieldMask:      Mask{"Original"},
			dstTranslation: Translation{Original: true, Language: language.English},
			srcTranslation: Translation{Original: false, Language: language.Latvian},
			want:           Translation{Original: false, Language: language.English},
		},
		{
			name:           "Update Original and Messages",
			fieldMask:      Mask{"Original", "Messages"},
			dstTranslation: Translation{Original: true, Language: language.English, Messages: []Message{{ID: "1", Message: "Hello"}}},  //nolint:lll
			srcTranslation: Translation{Original: false, Language: language.Latvian, Messages: []Message{{ID: "1", Message: "World"}}}, //nolint:lll
			want:           Translation{Original: false, Language: language.English, Messages: []Message{{ID: "1", Message: "World"}}}, //nolint:lll
		},
		{
			name:           "Update Messages",
			fieldMask:      Mask{"Messages"},
			dstTranslation: Translation{Messages: []Message{{ID: "1", Message: "Hello", Status: MessageStatusUntranslated}}},
			srcTranslation: Translation{Messages: []Message{{ID: "1", Message: "World", Status: MessageStatusTranslated}}},
			want:           Translation{Messages: []Message{{ID: "1", Message: "World", Status: MessageStatusTranslated}}},
		},
		{
			name:      "Update multiple Messages",
			fieldMask: Mask{"Messages"},
			dstTranslation: Translation{Messages: []Message{
				{ID: "1", Message: "Hello", Status: MessageStatusUntranslated},
				{ID: "2", Message: "Bonjour", Status: MessageStatusTranslated},
			}},
			srcTranslation: Translation{Messages: []Message{
				{ID: "2", Message: "Bonjour2", Status: MessageStatusUntranslated},
				{ID: "1", Message: "World", Status: MessageStatusTranslated},
			}},
			want: Translation{Messages: []Message{
				{ID: "1", Message: "World", Status: MessageStatusTranslated},
				{ID: "2", Message: "Bonjour2", Status: MessageStatusUntranslated},
			}},
		},
		{
			name:           "Update Messages, add new Message",
			fieldMask:      Mask{"Messages"},
			dstTranslation: Translation{Messages: []Message{{ID: "1", Message: "Hello"}}},
			srcTranslation: Translation{Messages: []Message{{ID: "2", Message: "World"}}},
			want:           Translation{Messages: []Message{{ID: "1", Message: "Hello"}, {ID: "2", Message: "World"}}},
		},
		{
			name:      "Update Messages, add multiple Messages",
			fieldMask: Mask{"Messages"},
			dstTranslation: Translation{Messages: []Message{
				{ID: "1", Message: "Hello"},
			}},
			srcTranslation: Translation{Messages: []Message{
				{ID: "2", Message: "World"},
				{ID: "3", Message: "Sun"},
				{ID: "4", Message: "Bye"},
			}},
			want: Translation{Messages: []Message{
				{ID: "1", Message: "Hello"},
				{ID: "2", Message: "World"},
				{ID: "3", Message: "Sun"},
				{ID: "4", Message: "Bye"},
			}},
		},
		{
			name:      "add Messages, update existing Messages in random order",
			fieldMask: Mask{"Messages"},
			dstTranslation: Translation{Messages: []Message{
				{ID: "1", Message: "Hello"},
			}},
			srcTranslation: Translation{Messages: []Message{
				{ID: "2", Message: "World"},
				{ID: "3", Message: "Sun"},
				{ID: "4", Message: "Bye"},
				{ID: "1", Message: "World"},
			}},
			want: Translation{Messages: []Message{
				{ID: "1", Message: "World"},
				{ID: "2", Message: "World"},
				{ID: "3", Message: "Sun"},
				{ID: "4", Message: "Bye"},
			}},
		},
		{
			name:      "Update multiple Messages, add new Message",
			fieldMask: Mask{"Messages"},
			dstTranslation: Translation{Messages: []Message{
				{ID: "1", Message: "Hello", Status: MessageStatusUntranslated},
				{ID: "2", Message: "Bonjour", Status: MessageStatusTranslated},
			}},
			srcTranslation: Translation{Messages: []Message{
				{ID: "1", Message: "World", Status: MessageStatusTranslated},
				{ID: "2", Message: "Bonjour2", Status: MessageStatusUntranslated},
				{ID: "3", Message: "Bye", Status: MessageStatusUntranslated},
			}},
			want: Translation{Messages: []Message{
				{ID: "1", Message: "World", Status: MessageStatusTranslated},
				{ID: "2", Message: "Bonjour2", Status: MessageStatusUntranslated},
				{ID: "3", Message: "Bye", Status: MessageStatusUntranslated},
			}},
		},
		{
			name:      "Update multiple Messages, add new Message in random order",
			fieldMask: Mask{"Messages"},
			dstTranslation: Translation{Messages: []Message{
				{ID: "1", Message: "Hello", Status: MessageStatusUntranslated},
				{ID: "2", Message: "Bonjour", Status: MessageStatusTranslated},
			}},
			srcTranslation: Translation{Messages: []Message{
				{ID: "3", Message: "Bye", Status: MessageStatusUntranslated},
				{ID: "1", Message: "World", Status: MessageStatusTranslated},
				{ID: "2", Message: "Bonjour2", Status: MessageStatusUntranslated},
			}},
			want: Translation{Messages: []Message{
				{ID: "1", Message: "World", Status: MessageStatusTranslated},
				{ID: "2", Message: "Bonjour2", Status: MessageStatusUntranslated},
				{ID: "3", Message: "Bye", Status: MessageStatusUntranslated},
			}},
		},
		{
			name:      "Add one Message to multiples",
			fieldMask: Mask{"Messages"},
			dstTranslation: Translation{Messages: []Message{
				{ID: "1", Message: "Hello", Status: MessageStatusUntranslated},
				{ID: "2", Message: "Bonjour", Status: MessageStatusTranslated},
			}},
			srcTranslation: Translation{Messages: []Message{
				{ID: "3", Message: "Bye", Status: MessageStatusUntranslated},
			}},
			want: Translation{Messages: []Message{
				{ID: "1", Message: "Hello", Status: MessageStatusUntranslated},
				{ID: "2", Message: "Bonjour", Status: MessageStatusTranslated},
				{ID: "3", Message: "Bye", Status: MessageStatusUntranslated},
			}},
		},
		{
			name:      "Update Message Description",
			fieldMask: Mask{"Messages"},
			dstTranslation: Translation{Original: true, Messages: []Message{
				{ID: "1", Message: "Hello", Description: "welcome"},
			}},
			srcTranslation: Translation{Original: true, Messages: []Message{
				{ID: "1", Message: "Hello", Description: "hi"},
			}},
			want: Translation{Original: true, Messages: []Message{
				{ID: "1", Message: "Hello", Description: "hi"},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			Update(&test.srcTranslation, &test.dstTranslation, test.fieldMask)

			if !reflect.DeepEqual(test.want, test.dstTranslation) {
				t.Errorf("want %v, got %v", test.want, test.dstTranslation)
			}
		})
	}
}
