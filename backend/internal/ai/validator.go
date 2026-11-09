package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"cloudprep-backend/internal/models"
)

// ComputeFingerprint generates the normalized SHA-256 fingerprint for deduplication.
func ComputeFingerprint(topicID, style, stem string) string {
	norm := strings.ToLower(strings.TrimSpace(topicID) + "|" + strings.TrimSpace(style) + "|" + strings.Join(strings.Fields(stem), " "))
	sum := sha256.Sum256([]byte(norm))
	return hex.EncodeToString(sum[:])
}

// NormalizeDifficulty maps various difficulty aliases to the canonical CloudPrep constants.
func NormalizeDifficulty(raw string) string {
	upper := strings.ToUpper(strings.TrimSpace(raw))
	switch upper {
	case "EASY", models.DifficultyFoundational:
		return models.DifficultyFoundational
	case "MEDIUM", models.DifficultyIntermediate:
		return models.DifficultyIntermediate
	case "HARD", models.DifficultyAdvanced:
		return models.DifficultyAdvanced
	default:
		return models.DifficultyIntermediate
	}
}

// NormalizeStyle maps style aliases to the canonical CloudPrep constants.
func NormalizeStyle(raw, fallback string) string {
	upper := strings.ToUpper(strings.TrimSpace(raw))
	switch upper {
	case models.StyleDefinition, models.StyleCommandSyntax, models.StyleComparison, models.StyleScenario:
		return upper
	default:
		if fallback != "" {
			return fallback
		}
		return models.StyleDefinition
	}
}

// ValidateQuestion verifies that an AI-generated question satisfies all quality and schema invariants.
func ValidateQuestion(q *GeneratedQuestion, topicID, defaultStyle string, existingFingerprints map[string]bool) ValidationResult {
	stem := strings.TrimSpace(q.Stem)
	if len(stem) < 15 {
		return ValidationResult{
			IsValid:         false,
			RejectionReason: "Stem is too short (must be >= 15 characters).",
		}
	}

	style := NormalizeStyle(q.QuestionStyle, defaultStyle)
	diff := NormalizeDifficulty(q.Difficulty)
	q.QuestionStyle = style
	q.Difficulty = diff

	// Options validation: must have exactly 4 distinct options
	if len(q.Options) != 4 {
		return ValidationResult{
			IsValid:         false,
			RejectionReason: fmt.Sprintf("Question must have exactly 4 options, got %d.", len(q.Options)),
		}
	}

	seenKeys := make(map[string]bool, 4)
	seenTexts := make(map[string]bool, 4)
	validKeys := map[string]bool{"A": true, "B": true, "C": true, "D": true}

	correctKey := strings.ToUpper(strings.TrimSpace(q.CorrectOptionKey))
	if !validKeys[correctKey] {
		return ValidationResult{
			IsValid:         false,
			RejectionReason: fmt.Sprintf("Invalid correctOptionKey %q (must be A, B, C, or D).", q.CorrectOptionKey),
		}
	}
	q.CorrectOptionKey = correctKey

	for i := range q.Options {
		opt := &q.Options[i]
		key := strings.ToUpper(strings.TrimSpace(opt.Key))
		text := strings.TrimSpace(opt.Text)

		if !validKeys[key] {
			return ValidationResult{
				IsValid:         false,
				RejectionReason: fmt.Sprintf("Invalid option key %q.", opt.Key),
			}
		}
		if seenKeys[key] {
			return ValidationResult{
				IsValid:         false,
				RejectionReason: fmt.Sprintf("Duplicate option key %q.", key),
			}
		}
		seenKeys[key] = true
		opt.Key = key

		if len(text) < 2 {
			return ValidationResult{
				IsValid:         false,
				RejectionReason: fmt.Sprintf("Option %s text is too short (%q).", key, text),
			}
		}

		normText := strings.ToLower(strings.Join(strings.Fields(text), " "))
		if seenTexts[normText] {
			return ValidationResult{
				IsValid:         false,
				RejectionReason: fmt.Sprintf("Duplicate option text across options in question (%q).", text),
			}
		}
		seenTexts[normText] = true
		opt.Text = text

		// Distractors must have a non-empty rationale
		if key != correctKey {
			rationale := strings.TrimSpace(opt.DistractorRationale)
			if len(rationale) < 5 {
				opt.DistractorRationale = fmt.Sprintf("Incorrect option: does not satisfy the requirements of %s.", topicID)
			}
		} else {
			opt.DistractorRationale = ""
		}
	}

	if len(seenKeys) != 4 {
		return ValidationResult{
			IsValid:         false,
			RejectionReason: "Question must have options labeled A, B, C, and D.",
		}
	}

	explanation := strings.TrimSpace(q.Explanation)
	if len(explanation) < 15 {
		return ValidationResult{
			IsValid:         false,
			RejectionReason: "Explanation must be at least 15 characters.",
		}
	}
	q.Explanation = explanation

	// Fingerprint and duplicate check
	fp := ComputeFingerprint(topicID, style, stem)
	if existingFingerprints != nil && existingFingerprints[fp] {
		return ValidationResult{
			IsValid:            false,
			ContentFingerprint: fp,
			RejectionReason:    "Duplicate question detected (matching content fingerprint in bank).",
		}
	}

	return ValidationResult{
		IsValid:            true,
		NormalizedStem:     stem,
		ContentFingerprint: fp,
	}
}
