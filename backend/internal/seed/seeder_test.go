package seed

import (
	"testing"

	"cloudprep-backend/internal/models"
)

func TestQuestionBankIntegrityAndDistribution(t *testing.T) {
	questions := BuildFullQuestionBank()

	if len(questions) < 1000 {
		t.Fatalf("expected >= 1,000 questions in seed bank, got %d", len(questions))
	}
	if len(questions) != 1040 {
		t.Fatalf("expected exactly 1,040 questions (104 per topic x 10 topics), got %d", len(questions))
	}

	seenIDs := make(map[string]bool, len(questions))
	seenFingerprints := make(map[string]string, len(questions))
	topicCounts := make(map[string]int)
	topicStyleCounts := make(map[string]map[string]int)
	difficultyCounts := make(map[string]int)

	for _, q := range questions {
		if seenIDs[q.ID] {
			t.Fatalf("duplicate question ID detected: %s", q.ID)
		}
		seenIDs[q.ID] = true

		if prevID, exists := seenFingerprints[q.ContentFingerprint]; exists {
			t.Fatalf("duplicate ContentFingerprint between %s and %s: %s", prevID, q.ID, q.Stem)
		}
		seenFingerprints[q.ContentFingerprint] = q.ID

		if q.Stem == "" || q.Explanation == "" || q.SourceReference == "" || q.ObjectiveID == "" {
			t.Fatalf("question %s is missing required metadata (stem/explanation/sourceReference/objectiveId)", q.ID)
		}

		if len(q.Options) != 4 {
			t.Fatalf("question %s expected 4 options, got %d", q.ID, len(q.Options))
		}

		optionTexts := make(map[string]bool)
		correctFound := false
		for _, opt := range q.Options {
			if opt.Text == "" {
				t.Fatalf("question %s option %s has empty text", q.ID, opt.Key)
			}
			if optionTexts[opt.Text] {
				t.Fatalf("question %s has duplicate option text: %q", q.ID, opt.Text)
			}
			optionTexts[opt.Text] = true

			if opt.Key == q.CorrectOptionKey {
				correctFound = true
			} else if opt.DistractorRationale == "" {
				t.Fatalf("question %s distractor %s is missing DistractorRationale", q.ID, opt.Key)
			}
		}
		if !correctFound {
			t.Fatalf("question %s CorrectOptionKey=%s not found in options", q.ID, q.CorrectOptionKey)
		}

		topicCounts[q.TopicID]++
		if _, ok := topicStyleCounts[q.TopicID]; !ok {
			topicStyleCounts[q.TopicID] = make(map[string]int)
		}
		topicStyleCounts[q.TopicID][q.QuestionStyle]++
		difficultyCounts[q.Difficulty]++
	}

	if len(topicCounts) != 10 {
		t.Fatalf("expected 10 topics in question bank, got %d", len(topicCounts))
	}

	for topicID, count := range topicCounts {
		if count < 100 {
			t.Errorf("topic %s has fewer than 100 questions: %d", topicID, count)
		}
		styles := topicStyleCounts[topicID]
		if styles[models.StyleDefinition] != 31 {
			t.Errorf("topic %s expected 31 DEFINITION questions, got %d", topicID, styles[models.StyleDefinition])
		}
		if styles[models.StyleCommandSyntax] != 26 {
			t.Errorf("topic %s expected 26 COMMAND_SYNTAX questions, got %d", topicID, styles[models.StyleCommandSyntax])
		}
		if styles[models.StyleComparison] != 21 {
			t.Errorf("topic %s expected 21 COMPARISON questions, got %d", topicID, styles[models.StyleComparison])
		}
		if styles[models.StyleScenario] != 26 {
			t.Errorf("topic %s expected 26 SCENARIO questions, got %d", topicID, styles[models.StyleScenario])
		}
	}

	if difficultyCounts[models.DifficultyFoundational] < 500 {
		t.Errorf("expected >= 500 questions for difficulty %s, got %d", models.DifficultyFoundational, difficultyCounts[models.DifficultyFoundational])
	}
	if difficultyCounts[models.DifficultyIntermediate] < 350 {
		t.Errorf("expected >= 350 questions for difficulty %s, got %d", models.DifficultyIntermediate, difficultyCounts[models.DifficultyIntermediate])
	}
	if difficultyCounts[models.DifficultyAdvanced] != 0 {
		t.Errorf("expected 0 questions for difficulty %s (strictly beginner & intermediate), got %d", models.DifficultyAdvanced, difficultyCounts[models.DifficultyAdvanced])
	}
}

