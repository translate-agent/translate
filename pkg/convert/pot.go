package convert

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	mf2 "go.expect.digital/mf2"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/pot"
)

const pluralCountLimit = 2 // max value for plural count. //TODO why ?

// ---------------------------------------PO->Translation---------------------------------------

// FromPo converts a byte slice representing a PO file to a model.Translation structure.
func FromPo(b []byte, originalOverride *bool) (model.Translation, error) {
	po, err := pot.Parse(bytes.NewReader(b))
	if err != nil {
		return model.Translation{}, fmt.Errorf("parse po file: %w", err)
	}

	translation := model.Translation{
		Language: po.Header.Language,
		Messages: make([]model.Message, 0, len(po.Messages)),
		Original: isOriginalPO(po, originalOverride),
	}

	var (
		getStatus   func(pot.MessageNode) model.MessageStatus // status getter based on originality
		getMessages func(pot.MessageNode) []string            // messages getter based on originality
	)

	switch translation.Original {
	// on original, all messages are considered translated
	// messages are get from msgid and msgid_plural if exists.
	case true:
		getStatus = func(pot.MessageNode) model.MessageStatus { return model.MessageStatusTranslated }
		getMessages = func(n pot.MessageNode) []string {
			if n.MsgIDPlural != "" {
				return []string{n.MsgID, n.MsgIDPlural}
			}

			return []string{n.MsgID}
		}
	// on non original, all messages are considered untranslated or fuzzy
	// messages are get from msgstr and msgstr[*] if exists.
	case false:
		getStatus = func(n pot.MessageNode) model.MessageStatus {
			if slices.Contains(n.Flags, "fuzzy") {
				return model.MessageStatusFuzzy
			}

			return model.MessageStatusUntranslated
		}
		getMessages = func(n pot.MessageNode) []string { return n.MsgStr }
	}

	for _, node := range po.Messages {
		mf2Msg, err := msgNodeToMF2(node, getMessages)
		if err != nil {
			return model.Translation{}, fmt.Errorf("convert message node to mf2 format: %w", err)
		}

		translation.Messages = append(translation.Messages, model.Message{
			ID:          node.MsgID,
			PluralID:    node.MsgIDPlural,
			Description: strings.Join(node.ExtractedComment, "\n"),
			Positions:   node.References,
			Message:     mf2Msg,
			Status:      getStatus(node),
		})
	}

	return translation, nil
}

// isOriginalPO function determines whether a PO file is an original or a translation.
func isOriginalPO(po pot.Po, override *bool) bool {
	// if override is not nil, use it.
	if override != nil {
		return *override
	}

	// if in all messages msgstr is empty, then it's original.
	allMsgStrEmpty := slices.ContainsFunc(po.Messages, func(node pot.MessageNode) bool {
		for _, msg := range node.MsgStr {
			if msg != "" {
				return false
			}
		}

		return true
	})

	// XXX: originality can also be determined by file extension.
	// .pot == original, .po == translation.
	// but we don't preserve file extension, so we can't use this method for now.

	return allMsgStrEmpty
}

// XXX: Can every convert use that ?
var placeholderFormats = map[string]*regexp.Regexp{
	"pythonVar":    regexp.MustCompile(`%\((\w+)\)([sd])`), // hello %(var)s | hello %(var)d
	"printf":       regexp.MustCompile(`%(s|d|f)`),         // hello %s | hello %d | hello %f
	"bracketVar":   regexp.MustCompile(`\{(\w+)\}`),        // hello {var} | hello {0}
	"emptyBracket": regexp.MustCompile(`\{\}`),             // hello {}
}

// msgNodeToMF2 function converts a pot.MessageNode to a MessageFormat2 string.
func msgNodeToMF2(node pot.MessageNode, getMessages func(pot.MessageNode) []string) (string, error) {
	// look for placeholder flag, e.g. #, python-format, c-format
	// avoid flags with "no-" prefix, e.g. #, no-python-format, no-c-format
	placeholderFlagIDx := slices.IndexFunc(node.Flags, func(flag string) bool {
		return strings.Contains(flag, "-format") && !strings.Contains(flag, "no-")
	})

	mfBuilder := mf2.NewBuilder()
	placeholders := make(map[string]struct{}) // map of placeholders to avoid duplicates, only for plural messages

	switch messages := getMessages(node); len(messages) {
	case 1: // singular message
		if placeholderFlagIDx == -1 { //  without placeholders
			mfBuilder.Text(messages[0])
		} else { // with placeholders
			mfBuilder.Local("format", mf2.Literal(node.Flags[placeholderFlagIDx])) // capture format flag
			if err := textWithPlaceholder(mfBuilder, messages[0], placeholders); err != nil {
				return "", fmt.Errorf("parse message with placeholders: %w", err)
			}
		}

	default: // plural message
		mfBuilder.Match(mf2.Var("$count")) // match to arbitrary variable name

		var build func(*mf2.Builder, string, map[string]struct{}) error // build function based on placeholders

		switch placeholderFlagIDx {
		case -1: // without placeholders
			build = func(mfBuilder *mf2.Builder, msg string, _ map[string]struct{}) error { mfBuilder.Text(msg); return nil }
		default: //  with placeholders
			mfBuilder.Local("format", mf2.Literal(node.Flags[placeholderFlagIDx])) // capture format flag

			build = textWithPlaceholder
		}

		for i := range messages {
			if i == len(messages)-1 {
				mfBuilder.Keys("*")
			} else {
				mfBuilder.Keys(i + 1) // arbitrary key names to match to //TODO can be derived from plural forms
			}

			if err := build(mfBuilder, messages[i], placeholders); err != nil {
				return "", fmt.Errorf("parse plural message with placeholders: %w", err)
			}
		}
	}

	mf2String, err := mfBuilder.Build()
	if err != nil {
		return "", fmt.Errorf("build mf2 string: %w", err)
	}

	return mf2String, nil
}

func textWithPlaceholder(mfBuilder *mf2.Builder, msg string, placeholders map[string]struct{}) error {
	for name, re := range placeholderFormats {
		var currentIdx int

		allIndices := re.FindAllStringIndex(msg, -1)
		for i, indices := range allIndices {
			// Add text before the placeholder
			text := msg[currentIdx:indices[0]]
			if text != "" {
				mfBuilder.Text(text)
			}

			originalVariable := msg[indices[0]:indices[1]]

			var mf2Variable string

			switch name {
			case "printf", "emptyBracket":
				mf2Variable = fmt.Sprintf("$ph%d", i) // %d|%s|%f -> $ph0|$ph1|$ph2...
			case "pythonVar":
				mf2Variable = "$" + originalVariable[2:len(originalVariable)-2] // %(var)s -> $var
			case "bracketVar":
				mf2Variable = "$" + originalVariable[1:len(originalVariable)-1] // {var} -> $var
			}

			// Avoid adding duplicate locals variables when working with plural messages
			if _, ok := placeholders[mf2Variable]; !ok {
				mfBuilder.Local(mf2Variable, mf2.Literal(originalVariable))

				placeholders[mf2Variable] = struct{}{}
			}

			mfBuilder.Expr(mf2.Var(mf2Variable))

			currentIdx = indices[1]

			// If this is the last placeholder, add the text after the placeholder
			if i == len(allIndices)-1 {
				mfBuilder.Text(msg[currentIdx:])
				return nil
			}
		}
	}

	return fmt.Errorf("format flag is present, but no placeholders found: '%s'", msg)
}

// ---------------------------------------Translation->PO---------------------------------------

type poTag string

const (
	MsgID        poTag = "msgid"
	PluralID     poTag = "msgid_plural"
	MsgStrPlural poTag = "msgstr[%d]"
	MsgStr       poTag = "msgstr"
)

// ToPot function takes a model.Translation structure,
// writes the language information and each message to a buffer in the POT file format,
// and returns the buffer contents as a byte slice representing the POT file.
// TODO instead of creating a buffer, create PO structure here and fill it with data.
// Then convert PO structure to bytes.
func ToPot(t model.Translation) ([]byte, error) {
	var b bytes.Buffer

	if _, err := fmt.Fprintf(&b, "msgid \"\"\nmsgstr \"\"\n\"Language: %s\\n\"\n", t.Language); err != nil {
		return nil, fmt.Errorf("write language: %w", err)
	}

	if !t.Original {
		// Temporary we support plural forms (one and other).
		// https://www.gnu.org/software/gettext/manual/html_node/Plural-forms.html
		if _, err := fmt.Fprintf(&b, "\"Plural-Forms: nplurals=%d; plural=(n != 1);\\n\"\n\n", pluralCountLimit); err != nil {
			return nil, fmt.Errorf("write Plural-Forms: %w", err)
		}
	}

	for i, message := range t.Messages {
		if err := writeMessage(&b, i, message); err != nil {
			return nil, fmt.Errorf("write message: %w", err)
		}
	}

	return b.Bytes(), nil
}

// writeToPoTag function selects the appropriate write function (writeDefault or writePlural)
// based on the tag type (poTag) and uses that function to write the tag value to a bytes.Buffer.
func writeToPoTag(b *bytes.Buffer, tag poTag, str string) error {
	write := writeDefault
	if tag == MsgStrPlural {
		write = writePlural
	}

	if err := write(b, tag, str); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// writeDefault parses a default message string into nodes, accumulates the text content,
// handles cases where the default message doesn't have any text, splits the accumulated text into lines,
// and writes the lines as a multiline tag value to a bytes.Buffer.
// TODO.
func writeDefault(b *bytes.Buffer, tag poTag, str string) error { return nil }

// INFO: old code for reference.
// func writeDefault(b *bytes.Buffer, tag poTag, str string) error {
// var text strings.Builder

// nodes, err := messageformat.Parse(str)
// if err != nil {
// 	return fmt.Errorf("parse message: %w", err)
// }

// for _, node := range nodes {
// 	nodeTxt, ok := node.(messageformat.NodeText)
// 	if !ok {
// 		return errors.New("convert node to messageformat.NodeText")
// 	}

// 	text.WriteString(nodeTxt.Text)
// }

// if text.String() == "" {
// 	text.WriteString(str)
// }

// lines := getPoTagLines(text.String())

// if len(lines) == 1 {
// 	if _, err = fmt.Fprintf(b, "%s \"%s\"\n", tag, lines[0]); err != nil {
// 		return fmt.Errorf("write %s: %w", tag, err)
// 	}

// 	return nil
// }

// if _, err = fmt.Fprintf(b, "%s \"\"\n", tag); err != nil {
// 	return fmt.Errorf("write %s: %w", tag, err)
// }

// if err = writeMultiline(b, tag, lines); err != nil {
// 	return fmt.Errorf("write multiline: %w", err)
// }

// return nil
//}

// writePlural parses a plural message string into nodes, iterates over the nodes,
// and writes the variants of the plural message to a bytes.Buffer.
// TODO.
func writePlural(b *bytes.Buffer, tag poTag, str string) error { return nil }

// INFO: old code for reference.
// func writePlural(b *bytes.Buffer, tag poTag, str string) error {
// nodes, err := messageformat.Parse(str)
// if err != nil {
// 	return fmt.Errorf("parse message: %w", err)
// }

// for _, node := range nodes {
// 	nodeMatch, ok := node.(messageformat.NodeMatch)
// 	if !ok {
// 		return errors.New("convert node to messageformat.NodeMatch")
// 	}

// 	if err = writeVariants(b, tag, nodeMatch); err != nil {
// 		return fmt.Errorf("write variants: %w", err)
// 	}
// }

// return nil
//}

// writeVariants writes the variants of a plural message to a bytes.Buffer.
// INFO: old code for reference.
// func writeVariants(b *bytes.Buffer, tag poTag, nodeMatch messageformat.NodeMatch) error {
// 	for i, variant := range nodeMatch.Variants {
// 		if _, err := fmt.Fprintf(b, "msgstr[%d] ", i); err != nil {
// 			return fmt.Errorf("write plural msgstr: %w", err)
// 		}

// 		var txt strings.Builder

// 		for _, msg := range variant.Message {
// 			switch node := msg.(type) {
// 			case messageformat.NodeText:
// 				txt.WriteString(node.Text)
// 			case messageformat.NodeVariable:
// 				txt.WriteString("%d")
// 			default:
// 				return errors.New("unknown node type")
// 			}
// 		}

// 		lines := getPoTagLines(txt.String())

// 		if len(lines) == 1 {
// 			if _, err := fmt.Fprintf(b, "\"%s\"\n", lines[0]); err != nil {
// 				return fmt.Errorf("write %s: %w", tag, err)
// 			}

// 			continue
// 		}

// 		if _, err := fmt.Fprintf(b, "\"\"\n"); err != nil {
// 			return fmt.Errorf("write %s: %w", tag, err)
// 		}

// 		if err := writeMultiline(b, tag, lines); err != nil {
// 			return fmt.Errorf("write multiline: %w", err)
// 		}
// 	}

// 	return nil
// }

// writeMultiline writes a slice of strings as a multiline tag value to a bytes.Buffer.
// TODO.
func writeMultiline(b *bytes.Buffer, tag poTag, lines []string) error { //nolint:unused
	for _, line := range lines {
		if !strings.HasSuffix(line, "\\n") {
			line += "\\n"
		}

		if _, err := fmt.Fprint(b, "\""+line+"\""+"\n"); err != nil {
			return fmt.Errorf("write %s: %w", tag, err)
		}
	}

	return nil
}

// getPoTagLines processes a string by encoding, splitting it into lines,
// and handling special cases where the last line is an empty string due to trailing newline characters.
// It returns a slice of strings representing the individual lines.
// TODO.
func getPoTagLines(str string) []string { //nolint:unused
	encodedStr := strconv.Quote(str)
	encodedStr = strings.ReplaceAll(encodedStr, "\\\\", "\\")
	encodedStr = encodedStr[1 : len(encodedStr)-1] // trim quotes
	lines := strings.Split(encodedStr, "\\n")

	// Remove the empty string element. The line is empty when splitting by "\\n" and "\n" is the last character in str.
	if lines[len(lines)-1] == "" {
		if len(lines) == 1 {
			return lines
		} else {
			lines = lines[:len(lines)-1]
			lines[len(lines)-1] += "\\n" // add the "\n" back to the last line
		}
	}

	return lines
}

// writeMessage writes a single message entry to a bytes.Buffer,
// including new lines, plural forms information, descriptions, fuzzy comment.
func writeMessage(b *bytes.Buffer, index int, message model.Message) error {
	if index > 0 {
		if _, err := fmt.Fprint(b, "\n"); err != nil {
			return fmt.Errorf("write new line: %w", err)
		}
	}

	descriptions := strings.Split(message.Description, "\n")

	for _, description := range descriptions {
		if description != "" {
			if _, err := fmt.Fprintf(b, "#. %s\n", description); err != nil {
				return fmt.Errorf("write description: %w", err)
			}
		}
	}

	for _, pos := range message.Positions {
		if _, err := fmt.Fprintf(b, "#: %s\n", pos); err != nil {
			return fmt.Errorf("write positions: %w", err)
		}
	}

	if message.Status == model.MessageStatusFuzzy {
		if _, err := fmt.Fprint(b, "#, fuzzy\n"); err != nil {
			return fmt.Errorf("write fuzzy: %w", err)
		}
	}

	return writeTags(b, message)
}

// writeTags writes specific tags (MsgId, MsgStr, PluralId, MsgStrPlural)
// along with their corresponding values to a bytes.Buffer.
func writeTags(b *bytes.Buffer, message model.Message) error {
	if err := writeToPoTag(b, MsgID, message.ID); err != nil {
		return fmt.Errorf("format msgid: %w", err)
	}

	// singular
	if message.PluralID == "" {
		if err := writeToPoTag(b, MsgStr, message.Message); err != nil {
			return fmt.Errorf("format msgstr: %w", err)
		}
	} else {
		// plural
		if err := writeToPoTag(b, PluralID, message.PluralID); err != nil {
			return fmt.Errorf("format msgid_plural: %w", err)
		}
		if err := writeToPoTag(b, MsgStrPlural, message.Message); err != nil {
			return fmt.Errorf("format msgstr[]: %w", err)
		}
	}

	return nil
}
