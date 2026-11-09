package seed

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"cloudprep-backend/internal/models"
)

// ConceptFact represents a core curriculum concept/term grounded in a specific course slide deck.
type ConceptFact struct {
	ObjectiveID   string
	Term          string
	Definition    string
	KeyMechanism  string
	ExampleUse    string
	CommonMistake string
}

// CommandFact represents a CLI command, Bash snippet, Git command, IAM JSON snippet, or cloud service CLI/config item.
type CommandFact struct {
	ObjectiveID  string
	CommandOrSyn string
	Purpose      string
	OutputOrBhvr string
	WrongFlagAlt string
	WrongWhy     string
}

// ComparisonPair represents two closely related curriculum concepts that students must distinguish on exams.
type ComparisonPair struct {
	ObjectiveID string
	ConceptA    string
	ConceptB    string
	TraitA      string
	TraitB      string
	Dimension   string
	WhenToUseA  string
	WhenToUseB  string
}

// ScenarioCase represents an applied architecture, security, cost, or troubleshooting case grounded in the course notes.
type ScenarioCase struct {
	ObjectiveID     string
	ScenarioType    string
	Context         string
	Requirement     string
	BestSolution    string
	WhyBest         string
	BadAlternative1 string
	WhyBad1         string
	BadAlternative2 string
	WhyBad2         string
	BadAlternative3 string
	WhyBad3         string
}

// TopicBlueprint holds the curriculum facts for a single topic used to construct >=104 unique multi-style questions.
type TopicBlueprint struct {
	TopicID         string
	TopicTitle      string
	SourceReference string
	Concepts        []ConceptFact
	Commands        []CommandFact
	Comparisons     []ComparisonPair
	Scenarios       []ScenarioCase
}

// ComputeFingerprint generates a normalized SHA-256 content hash for deduplication.
func ComputeFingerprint(topicID, style, stem string) string {
	norm := strings.ToLower(strings.TrimSpace(topicID) + "|" + strings.TrimSpace(style) + "|" + strings.Join(strings.Fields(stem), " "))
	sum := sha256.Sum256([]byte(norm))
	return hex.EncodeToString(sum[:])
}

func stripMarkdown(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "*\"", "\"")
	s = strings.ReplaceAll(s, "\"*", "\"")
	s = strings.ReplaceAll(s, "> *", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.TrimPrefix(s, "- ")
	s = strings.TrimPrefix(s, "* ")
	s = strings.TrimPrefix(s, "> ")
	s = strings.TrimPrefix(s, "# ")

	// If multiline snippet (e.g. bash scripts), convert newlines to proper command separators ';'
	lines := strings.Split(s, "\n")
	cleanLines := make([]string, 0, len(lines))
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}
	if len(cleanLines) > 1 {
		var sb strings.Builder
		for idx, l := range cleanLines {
			if idx > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(l)
			if idx < len(cleanLines)-1 {
				if !strings.HasSuffix(l, ";") && !strings.HasSuffix(l, "then") && !strings.HasSuffix(l, "do") && !strings.HasSuffix(l, "else") && !strings.HasSuffix(l, "{") {
					sb.WriteString(";")
				}
			}
		}
		return sb.String()
	}
	return strings.Join(strings.Fields(s), " ")
}

func cleanTerm(s string) string {
	s = stripMarkdown(s)
	// Remove trailing notes or clean up
	return strings.TrimSpace(s)
}

func cleanSentence(s string) string {
	s = stripMarkdown(s)
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "?") && !strings.HasSuffix(s, "!") {
		s += "."
	}
	return s
}

func cleanPurpose(s string) string {
	s = stripMarkdown(s)
	s = strings.TrimSpace(s)
	if len(s) > 0 {
		first := strings.ToLower(string(s[0]))
		s = first + s[1:]
	}
	return strings.TrimSuffix(s, ".")
}

func cleanOutput(s string) string {
	s = stripMarkdown(s)
	s = strings.TrimSpace(s)
	return cleanSentence(s)
}

func cleanTrait(s string) string {
	s = stripMarkdown(s)
	s = strings.TrimSpace(s)
	if len(s) > 0 {
		first := strings.ToLower(string(s[0]))
		s = first + s[1:]
	}
	return strings.TrimSuffix(s, ".")
}

func difficultyForIndex(idx int) (string, string) {
	// Strictly Beginner (60% Foundational) and Intermediate (40% Intermediate), 0% Advanced
	switch idx % 5 {
	case 0, 1, 2:
		return models.DifficultyFoundational, "RECALL"
	default:
		return models.DifficultyIntermediate, "COMPREHENSION"
	}
}

func rotateCorrectKey(idx int) string {
	keys := []string{"A", "B", "C", "D"}
	return keys[idx%4]
}

// placeOptions places the correct option at targetCorrectKey and distributes the 3 distractors across the remaining keys.
func placeOptions(questionID, targetCorrectKey, correctText string, distractors [3]struct {
	Text      string
	Rationale string
}) []models.QuestionOption {
	keys := []string{"A", "B", "C", "D"}
	options := make([]models.QuestionOption, 0, 4)
	dIdx := 0
	for order, k := range keys {
		if k == targetCorrectKey {
			options = append(options, models.QuestionOption{
				ID:                  fmt.Sprintf("%s-opt-%s", questionID, k),
				QuestionID:          questionID,
				Key:                 k,
				Text:                correctText,
				DistractorRationale: "",
				Ordering:            order + 1,
			})
		} else {
			d := distractors[dIdx]
			dIdx++
			options = append(options, models.QuestionOption{
				ID:                  fmt.Sprintf("%s-opt-%s", questionID, k),
				QuestionID:          questionID,
				Key:                 k,
				Text:                d.Text,
				DistractorRationale: d.Rationale,
				Ordering:            order + 1,
			})
		}
	}
	return options
}

// BuildTopicQuestions constructs 104 unique, curriculum-grounded questions for a TopicBlueprint:
//   - 31 DEFINITION (~30%)
//   - 26 COMMAND_SYNTAX (~25%)
//   - 21 COMPARISON (~20%)
//   - 26 SCENARIO (~25%)
//
// Total = 104 questions per topic (1,040 across all 10 topics).
func BuildTopicQuestions(bp TopicBlueprint, now time.Time) []models.Question {
	questions := make([]models.Question, 0, 104)
	qCounter := 0
	nConcepts := len(bp.Concepts)
	nCommands := len(bp.Commands)
	nComparisons := len(bp.Comparisons)
	nScenarios := len(bp.Scenarios)

	// 1. Build 31 DEFINITION questions using 4 distinct angles over bp.Concepts
	defTemplates := []func(c, d1, d2, d3 ConceptFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }){
		// Angle 0: Direct definition of Term
		func(c, d1, d2, d3 ConceptFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("What is the primary function of %s?", cleanTerm(c.Term))
			correct := cleanSentence(c.Definition)
			explanation := fmt.Sprintf("%s is defined as: %s Key mechanism: %s.", cleanTerm(c.Term), cleanSentence(c.Definition), cleanSentence(c.KeyMechanism))
			distractors := [3]struct{ Text, Rationale string }{
				{cleanSentence(d1.Definition), fmt.Sprintf("This describes %s, not %s.", cleanTerm(d1.Term), cleanTerm(c.Term))},
				{cleanSentence(d2.Definition), fmt.Sprintf("This defines %s rather than %s.", cleanTerm(d2.Term), cleanTerm(c.Term))},
				{cleanSentence(d3.Definition), fmt.Sprintf("This refers to %s.", cleanTerm(d3.Term))},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 1: Identify Term from its clear plain-text definition
		func(c, d1, d2, d3 ConceptFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("Which concept matches this description: %s", cleanSentence(c.Definition))
			correct := cleanTerm(c.Term)
			explanation := fmt.Sprintf("%s matches this description: %s (Mechanism: %s).", cleanTerm(c.Term), cleanSentence(c.Definition), cleanSentence(c.KeyMechanism))
			distractors := [3]struct{ Text, Rationale string }{
				{cleanTerm(d1.Term), fmt.Sprintf("%s serves a different purpose: %s.", cleanTerm(d1.Term), cleanSentence(d1.Definition))},
				{cleanTerm(d2.Term), fmt.Sprintf("%s is incorrect; its role is: %s.", cleanTerm(d2.Term), cleanSentence(d2.Definition))},
				{cleanTerm(d3.Term), fmt.Sprintf("%s is incorrect; it is used for: %s.", cleanTerm(d3.Term), cleanSentence(d3.Definition))},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 2: Practical workflow task matching (clean, friendly concept identification)
		func(c, d1, d2, d3 ConceptFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("Which tool or concept is commonly used for: %s?", cleanSentence(c.ExampleUse))
			correct := cleanTerm(c.Term)
			explanation := fmt.Sprintf("%s is commonly used for this task: %s.", cleanTerm(c.Term), cleanSentence(c.ExampleUse))
			distractors := [3]struct{ Text, Rationale string }{
				{cleanTerm(d1.Term), fmt.Sprintf("%s is used for: %s.", cleanTerm(d1.Term), cleanSentence(d1.ExampleUse))},
				{cleanTerm(d2.Term), fmt.Sprintf("%s is used for: %s.", cleanTerm(d2.Term), cleanSentence(d2.ExampleUse))},
				{cleanTerm(d3.Term), fmt.Sprintf("%s is used for: %s.", cleanTerm(d3.Term), cleanSentence(d3.ExampleUse))},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 3: Practical implementation & use case
		func(c, d1, d2, d3 ConceptFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("In cloud and DevOps environments, what is %s commonly used for?", cleanTerm(c.Term))
			correct := cleanSentence(c.ExampleUse)
			explanation := fmt.Sprintf("%s is used for: %s.", cleanTerm(c.Term), cleanSentence(c.ExampleUse))
			distractors := [3]struct{ Text, Rationale string }{
				{cleanSentence(d1.ExampleUse), fmt.Sprintf("This applies to %s rather than %s.", cleanTerm(d1.Term), cleanTerm(c.Term))},
				{cleanSentence(d2.ExampleUse), fmt.Sprintf("This applies to %s rather than %s.", cleanTerm(d2.Term), cleanTerm(c.Term))},
				{cleanSentence(d3.ExampleUse), fmt.Sprintf("This describes %s.", cleanTerm(d3.Term))},
			}
			return stem, correct, explanation, distractors
		},
	}

	for i := 0; i < 31; i++ {
		qCounter++
		c := bp.Concepts[i%nConcepts]
		d1 := bp.Concepts[(i+1)%nConcepts]
		d2 := bp.Concepts[(i+3)%nConcepts]
		d3 := bp.Concepts[(i+5)%nConcepts]
		angle := (i / nConcepts) % len(defTemplates)
		stem, correct, explanation, distractors := defTemplates[angle](c, d1, d2, d3, i)

		qID := fmt.Sprintf("q-%s-def-%03d", bp.TopicID, i+1)
		diff, skill := difficultyForIndex(qCounter)
		targetKey := rotateCorrectKey(qCounter)
		opts := placeOptions(qID, targetKey, correct, distractors)

		questions = append(questions, models.Question{
			ID:                 qID,
			TopicID:            bp.TopicID,
			ObjectiveID:        c.ObjectiveID,
			Stem:               stem,
			QuestionStyle:      models.StyleDefinition,
			Difficulty:         diff,
			CognitiveSkill:     skill,
			CorrectOptionKey:   targetKey,
			Explanation:        explanation,
			SourceType:         models.SourceSeedBank,
			SourceReference:    bp.SourceReference,
			ContentFingerprint: ComputeFingerprint(bp.TopicID, models.StyleDefinition, stem),
			Status:             models.QuestionStatusApproved,
			Version:            1,
			Options:            opts,
			CreatedAt:          now,
			UpdatedAt:          now,
		})
	}

	// 2. Build 26 COMMAND_SYNTAX questions using 4 distinct angles over bp.Commands
	cmdTemplates := []func(cmd, d1, d2, d3 CommandFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }){
		// Angle 0: Select the exact command/syntax for a task
		func(cmd, d1, d2, d3 CommandFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("Which command is used to %s?", cleanPurpose(cmd.Purpose))
			correct := stripMarkdown(cmd.CommandOrSyn)
			explanation := fmt.Sprintf("The command %s is used to %s.", stripMarkdown(cmd.CommandOrSyn), cleanPurpose(cmd.Purpose))
			distractors := [3]struct{ Text, Rationale string }{
				{stripMarkdown(cmd.WrongFlagAlt), cleanSentence(cmd.WrongWhy)},
				{stripMarkdown(d1.CommandOrSyn), fmt.Sprintf("%s is used to %s.", stripMarkdown(d1.CommandOrSyn), cleanPurpose(d1.Purpose))},
				{stripMarkdown(d2.CommandOrSyn), fmt.Sprintf("%s is used to %s.", stripMarkdown(d2.CommandOrSyn), cleanPurpose(d2.Purpose))},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 1: Interpret the behavior/output of a command
		func(cmd, d1, d2, d3 CommandFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("What is the command %s used for?", stripMarkdown(cmd.CommandOrSyn))
			correct := cleanSentence(cmd.Purpose)
			explanation := fmt.Sprintf("Running %s accomplishes: %s (%s).", stripMarkdown(cmd.CommandOrSyn), cleanSentence(cmd.Purpose), cleanSentence(cmd.OutputOrBhvr))
			distractors := [3]struct{ Text, Rationale string }{
				{cleanSentence(d1.Purpose), fmt.Sprintf("This describes the role of %s.", stripMarkdown(d1.CommandOrSyn))},
				{cleanSentence(d2.Purpose), fmt.Sprintf("This describes the role of %s.", stripMarkdown(d2.CommandOrSyn))},
				{cleanSentence(d3.Purpose), fmt.Sprintf("This describes the role of %s.", stripMarkdown(d3.CommandOrSyn))},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 2: Diagnose why an incorrect syntax fails and select the correct replacement
		func(cmd, d1, d2, d3 CommandFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("Which command is the correct syntax to %s?", cleanPurpose(cmd.Purpose))
			correct := stripMarkdown(cmd.CommandOrSyn)
			explanation := fmt.Sprintf("%s is the standard syntax to %s.", stripMarkdown(cmd.CommandOrSyn), cleanPurpose(cmd.Purpose))
			distractors := [3]struct{ Text, Rationale string }{
				{stripMarkdown(cmd.WrongFlagAlt), cleanSentence(cmd.WrongWhy)},
				{stripMarkdown(d1.CommandOrSyn), fmt.Sprintf("%s is used to %s.", stripMarkdown(d1.CommandOrSyn), cleanPurpose(d1.Purpose))},
				{stripMarkdown(d2.CommandOrSyn), fmt.Sprintf("%s is used to %s.", stripMarkdown(d2.CommandOrSyn), cleanPurpose(d2.Purpose))},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 3: Match expected runtime behavior/output back to the exact command (clean, no blockquotes)
		func(cmd, d1, d2, d3 CommandFact, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("Which command produces this outcome: %s?", cleanSentence(cmd.OutputOrBhvr))
			correct := stripMarkdown(cmd.CommandOrSyn)
			explanation := fmt.Sprintf("%s produces this outcome: %s (%s).", stripMarkdown(cmd.CommandOrSyn), cleanSentence(cmd.OutputOrBhvr), cleanPurpose(cmd.Purpose))
			distractors := [3]struct{ Text, Rationale string }{
				{stripMarkdown(cmd.WrongFlagAlt), cleanSentence(cmd.WrongWhy)},
				{stripMarkdown(d1.CommandOrSyn), fmt.Sprintf("%s produces: %s.", stripMarkdown(d1.CommandOrSyn), cleanSentence(d1.OutputOrBhvr))},
				{stripMarkdown(d3.CommandOrSyn), fmt.Sprintf("%s produces: %s.", stripMarkdown(d3.CommandOrSyn), cleanSentence(d3.OutputOrBhvr))},
			}
			return stem, correct, explanation, distractors
		},
	}

	for i := 0; i < 26; i++ {
		qCounter++
		cmd := bp.Commands[i%nCommands]
		d1 := bp.Commands[(i+1)%nCommands]
		d2 := bp.Commands[(i+2)%nCommands]
		d3 := bp.Commands[(i+4)%nCommands]
		angle := (i / nCommands) % len(cmdTemplates)
		stem, correct, explanation, distractors := cmdTemplates[angle](cmd, d1, d2, d3, i)

		qID := fmt.Sprintf("q-%s-cmd-%03d", bp.TopicID, i+1)
		diff, skill := difficultyForIndex(qCounter)
		targetKey := rotateCorrectKey(qCounter)
		opts := placeOptions(qID, targetKey, correct, distractors)

		questions = append(questions, models.Question{
			ID:                 qID,
			TopicID:            bp.TopicID,
			ObjectiveID:        cmd.ObjectiveID,
			Stem:               stem,
			QuestionStyle:      models.StyleCommandSyntax,
			Difficulty:         diff,
			CognitiveSkill:     skill,
			CorrectOptionKey:   targetKey,
			Explanation:        explanation,
			SourceType:         models.SourceSeedBank,
			SourceReference:    bp.SourceReference,
			ContentFingerprint: ComputeFingerprint(bp.TopicID, models.StyleCommandSyntax, stem),
			Status:             models.QuestionStatusApproved,
			Version:            1,
			Options:            opts,
			CreatedAt:          now,
			UpdatedAt:          now,
		})
	}

	// 3. Build 21 COMPARISON questions using 4 distinct angles over bp.Comparisons
	cmpTemplates := []func(p, d1, d2 ComparisonPair, idx int) (string, string, string, [3]struct{ Text, Rationale string }){
		// Angle 0: Direct contrast between Concept A and Concept B (NO ASTERISKS, NO RAW JARGON CATEGORIES)
		func(p, d1, d2 ComparisonPair, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("What is the main difference between %s and %s?", cleanTerm(p.ConceptA), cleanTerm(p.ConceptB))
			correct := fmt.Sprintf("%s %s, while %s %s.", cleanTerm(p.ConceptA), cleanTrait(p.TraitA), cleanTerm(p.ConceptB), cleanTrait(p.TraitB))
			explanation := fmt.Sprintf("%s exhibits %s, whereas %s exhibits %s.", cleanTerm(p.ConceptA), cleanSentence(p.TraitA), cleanTerm(p.ConceptB), cleanSentence(p.TraitB))
			distractors := [3]struct{ Text, Rationale string }{
				{fmt.Sprintf("%s %s, while %s %s.", cleanTerm(p.ConceptA), cleanTrait(p.TraitB), cleanTerm(p.ConceptB), cleanTrait(p.TraitA)), fmt.Sprintf("This reverses the characteristics of %s and %s.", cleanTerm(p.ConceptA), cleanTerm(p.ConceptB))},
				{fmt.Sprintf("Both %s and %s have identical functionality.", cleanTerm(p.ConceptA), cleanTerm(p.ConceptB)), fmt.Sprintf("Incorrect: %s and %s differ fundamentally.", cleanTerm(p.ConceptA), cleanTerm(p.ConceptB))},
				{fmt.Sprintf("%s is only used on physical on-premises servers, while %s cannot run on Linux.", cleanTerm(p.ConceptA), cleanTerm(p.ConceptB)), "Incorrect distractor."},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 1: When to choose Concept A
		func(p, d1, d2 ComparisonPair, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("When should you choose %s instead of %s?", cleanTerm(p.ConceptA), cleanTerm(p.ConceptB))
			correct := fmt.Sprintf("When the requirement is %s.", cleanTrait(p.WhenToUseA))
			explanation := fmt.Sprintf("%s is optimal when %s, whereas %s is for %s.", cleanTerm(p.ConceptA), cleanSentence(p.WhenToUseA), cleanTerm(p.ConceptB), cleanSentence(p.WhenToUseB))
			distractors := [3]struct{ Text, Rationale string }{
				{fmt.Sprintf("When the requirement is %s.", cleanTrait(p.WhenToUseB)), fmt.Sprintf("This describes when to use %s.", cleanTerm(p.ConceptB))},
				{fmt.Sprintf("When the requirement is %s.", cleanTrait(d1.WhenToUseA)), fmt.Sprintf("This applies to %s.", cleanTerm(d1.ConceptA))},
				{"Only for legacy on-premises workloads.", fmt.Sprintf("%s is widely used in modern cloud deployments.", cleanTerm(p.ConceptA))},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 2: When to choose Concept B
		func(p, d1, d2 ComparisonPair, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("When should you choose %s instead of %s?", cleanTerm(p.ConceptB), cleanTerm(p.ConceptA))
			correct := fmt.Sprintf("When the requirement is %s.", cleanTrait(p.WhenToUseB))
			explanation := fmt.Sprintf("%s is preferred when %s, whereas %s is suited for %s.", cleanTerm(p.ConceptB), cleanSentence(p.WhenToUseB), cleanTerm(p.ConceptA), cleanSentence(p.WhenToUseA))
			distractors := [3]struct{ Text, Rationale string }{
				{fmt.Sprintf("When the requirement is %s.", cleanTrait(p.WhenToUseA)), fmt.Sprintf("This describes when to use %s.", cleanTerm(p.ConceptA))},
				{fmt.Sprintf("When the requirement is %s.", cleanTrait(d2.WhenToUseB)), fmt.Sprintf("This applies to %s.", cleanTerm(d2.ConceptB))},
				{"Only for single-node local development.", fmt.Sprintf("%s is a production-grade cloud solution.", cleanTerm(p.ConceptB))},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 3: Migration / trade-off effect
		func(p, d1, d2 ComparisonPair, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("What is the main difference when switching from %s to %s?", cleanTerm(p.ConceptA), cleanTerm(p.ConceptB))
			correct := fmt.Sprintf("Changes from %s to %s.", cleanTrait(p.TraitA), cleanTrait(p.TraitB))
			explanation := fmt.Sprintf("Switching from %s to %s shifts from %s to %s.", cleanTerm(p.ConceptA), cleanTerm(p.ConceptB), cleanSentence(p.TraitA), cleanSentence(p.TraitB))
			distractors := [3]struct{ Text, Rationale string }{
				{fmt.Sprintf("Changes from %s to %s.", cleanTrait(p.TraitB), cleanTrait(p.TraitA)), "This reverses the direction of the switch."},
				{"There is no functional difference between them.", "The two options differ fundamentally in how they work."},
				{"It completely eliminates all cloud services.", "Both options are standard cloud technologies."},
			}
			return stem, correct, explanation, distractors
		},
	}

	for i := 0; i < 21; i++ {
		qCounter++
		p := bp.Comparisons[i%nComparisons]
		d1 := bp.Comparisons[(i+1)%nComparisons]
		d2 := bp.Comparisons[(i+2)%nComparisons]
		angle := (i / nComparisons) % len(cmpTemplates)
		stem, correct, explanation, distractors := cmpTemplates[angle](p, d1, d2, i)

		qID := fmt.Sprintf("q-%s-cmp-%03d", bp.TopicID, i+1)
		diff, skill := difficultyForIndex(qCounter)
		targetKey := rotateCorrectKey(qCounter)
		opts := placeOptions(qID, targetKey, correct, distractors)

		questions = append(questions, models.Question{
			ID:                 qID,
			TopicID:            bp.TopicID,
			ObjectiveID:        p.ObjectiveID,
			Stem:               stem,
			QuestionStyle:      models.StyleComparison,
			Difficulty:         diff,
			CognitiveSkill:     skill,
			CorrectOptionKey:   targetKey,
			Explanation:        explanation,
			SourceType:         models.SourceSeedBank,
			SourceReference:    bp.SourceReference,
			ContentFingerprint: ComputeFingerprint(bp.TopicID, models.StyleComparison, stem),
			Status:             models.QuestionStatusApproved,
			Version:            1,
			Options:            opts,
			CreatedAt:          now,
			UpdatedAt:          now,
		})
	}

	// 4. Build 26 SCENARIO questions using 4 distinct angles over bp.Scenarios
	scnTemplates := []func(s ScenarioCase, idx int) (string, string, string, [3]struct{ Text, Rationale string }){
		// Angle 0: Primary practical solution recommendation
		func(s ScenarioCase, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("%s\n\nRequirement: %s\n\nWhich solution best meets this requirement?", stripMarkdown(s.Context), stripMarkdown(s.Requirement))
			correct := cleanTerm(s.BestSolution)
			explanation := fmt.Sprintf("%s is the optimal choice because: %s.", cleanTerm(s.BestSolution), cleanSentence(s.WhyBest))
			distractors := [3]struct{ Text, Rationale string }{
				{cleanTerm(s.BadAlternative1), cleanSentence(s.WhyBad1)},
				{cleanTerm(s.BadAlternative2), cleanSentence(s.WhyBad2)},
				{cleanTerm(s.BadAlternative3), cleanSentence(s.WhyBad3)},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 1: Practical reason for choice
		func(s ScenarioCase, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("%s\n\nThe team decides to use %s.\n\nWhat is the main reason for this choice?", stripMarkdown(s.Context), cleanTerm(s.BestSolution))
			correct := cleanSentence(s.WhyBest)
			explanation := fmt.Sprintf("Selecting %s is correct because: %s.", cleanTerm(s.BestSolution), cleanSentence(s.WhyBest))
			distractors := [3]struct{ Text, Rationale string }{
				{cleanSentence(s.WhyBad1), fmt.Sprintf("This alternative has drawbacks: %s", cleanSentence(s.WhyBad1))},
				{cleanSentence(s.WhyBad2), fmt.Sprintf("This alternative is problematic: %s", cleanSentence(s.WhyBad2))},
				{cleanSentence(s.WhyBad3), fmt.Sprintf("This approach does not fit: %s", cleanSentence(s.WhyBad3))},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 2: Why an alternative is not recommended
		func(s ScenarioCase, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("%s\n\nA colleague suggests using %s to %s.\n\nWhy is this not the recommended approach?", stripMarkdown(s.Context), cleanTerm(s.BadAlternative1), cleanPurpose(s.Requirement))
			correct := cleanSentence(s.WhyBad1)
			explanation := fmt.Sprintf("That proposal is not recommended because: %s. The preferred solution is %s (%s).", cleanSentence(s.WhyBad1), cleanTerm(s.BestSolution), cleanSentence(s.WhyBest))
			distractors := [3]struct{ Text, Rationale string }{
				{cleanSentence(s.WhyBad2), "This applies to a different scenario."},
				{cleanSentence(s.WhyBad3), "This is an unrelated concern."},
				{"It introduces unnecessary complexity without solving the problem.", "This approach does not meet cloud engineering best practices."},
			}
			return stem, correct, explanation, distractors
		},
		// Angle 3: Practical resolution
		func(s ScenarioCase, idx int) (string, string, string, [3]struct{ Text, Rationale string }) {
			stem := fmt.Sprintf("%s\n\nAn attempt with %s failed. What is the recommended solution to %s?", stripMarkdown(s.Context), cleanTerm(s.BadAlternative2), cleanPurpose(s.Requirement))
			correct := cleanTerm(s.BestSolution)
			explanation := fmt.Sprintf("Using %s resolves this issue: %s.", cleanTerm(s.BestSolution), cleanSentence(s.WhyBest))
			distractors := [3]struct{ Text, Rationale string }{
				{cleanTerm(s.BadAlternative1), cleanSentence(s.WhyBad1)},
				{cleanTerm(s.BadAlternative3), cleanSentence(s.WhyBad3)},
				{"Disable monitoring and alerts to ignore the failure.", "Ignoring failures leads to unmonitored outages."},
			}
			return stem, correct, explanation, distractors
		},
	}

	for i := 0; i < 26; i++ {
		qCounter++
		s := bp.Scenarios[i%nScenarios]
		angle := (i / nScenarios) % len(scnTemplates)
		stem, correct, explanation, distractors := scnTemplates[angle](s, i)

		qID := fmt.Sprintf("q-%s-scn-%03d", bp.TopicID, i+1)
		diff, skill := difficultyForIndex(qCounter)
		targetKey := rotateCorrectKey(qCounter)
		opts := placeOptions(qID, targetKey, correct, distractors)

		questions = append(questions, models.Question{
			ID:                 qID,
			TopicID:            bp.TopicID,
			ObjectiveID:        s.ObjectiveID,
			Stem:               stem,
			QuestionStyle:      models.StyleScenario,
			Difficulty:         diff,
			CognitiveSkill:     skill,
			ScenarioType:       s.ScenarioType,
			CorrectOptionKey:   targetKey,
			Explanation:        explanation,
			SourceType:         models.SourceSeedBank,
			SourceReference:    bp.SourceReference,
			ContentFingerprint: ComputeFingerprint(bp.TopicID, models.StyleScenario, stem),
			Status:             models.QuestionStatusApproved,
			Version:            1,
			Options:            opts,
			CreatedAt:          now,
			UpdatedAt:          now,
		})
	}

	return questions
}
