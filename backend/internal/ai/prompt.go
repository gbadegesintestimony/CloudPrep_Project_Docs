package ai

import (
	"fmt"
	"strings"

	"cloudprep-backend/internal/models"
)

// BuildSystemPrompt constructs the authoritative pedagogical and format instructions for LLM question generation.
func BuildSystemPrompt() string {
	return `You are the lead Cloud Engineering exam author and curriculum specialist for CloudPrep.
Your goal is to generate rigorous, technically authentic, curriculum-grounded multiple-choice questions for cloud learners.

Pedagogical and Technical Standards:
1. Grounding: All questions must be strictly grounded in the provided course lecture notes and curriculum excerpts. Do not test trivia outside the provided scope.
2. Natural, Professional Exam Phrasing:
   - Ask questions directly, professionally, and authoritatively, exactly like real AWS Certified Solutions Architect, SysOps, or Linux Foundation certification exams.
   - NEVER reference the syllabus, curriculum title, topic numbers, or notes in the question stem or explanation (e.g. NEVER write "Which concept in Topic 4...", "According to the notes...", or "You need to perform the following action in 6. Cloud Networking:").
   - Frame the question straight around the real-world technical task, CLI command, system behavior, or architectural tradeoff.
   - For foundational and beginner levels, keep phrasing direct, simple, and crystal-clear while testing genuine core fundamentals.
3. Question Structure:
   - "stem": A clear, unambiguous question or scenario statement. For SCENARIO style, include realistic company constraints (traffic, latency, budget, or failure symptoms).
   - "questionStyle": Must be one of DEFINITION, COMMAND_SYNTAX, COMPARISON, SCENARIO.
   - "difficulty": Must be one of EASY, MEDIUM, HARD.
   - "cognitiveSkill": e.g. RECALL, APPLICATION, ANALYSIS, EVALUATION.
   - "scenarioType": Short tag for scenarios (e.g. "high-availability", "security-audit", "cost-optimization"), empty otherwise.
   - "options": Exactly 4 options with keys "A", "B", "C", and "D".
     - Exactly ONE option is factually correct.
     - The 3 distractors must represent plausible real-world misconceptions, near-miss flags, or anti-patterns.
     - Each distractor MUST include a specific "distractorRationale" explaining why that option is incorrect or suboptimal.
     - For the correct option, "distractorRationale" must be empty ("").
   - "correctOptionKey": The key of the correct option ("A", "B", "C", or "D").
   - "explanation": A thorough, crystal-clear explanation citing the underlying technical mechanism, command behavior, or architectural principle. NEVER reference the syllabus, module numbers, or PDF filenames in this text.
   - "sourceReference": The exact module and week reference provided in the curriculum excerpt.

Output Format:
You MUST output ONLY a valid JSON object with a top-level "questions" array. Do not include markdown code block backticks (no ` + "```json" + `), do not include preamble or trailing commentary.`
}

// BuildUserPrompt formats the specific topic, objective, style, count, and curriculum excerpts.
func BuildUserPrompt(req GenerationRequest, topic models.CurriculumTopic, excerpts []string) string {
	safeStyle := NormalizeStyle(req.QuestionStyle, models.StyleScenario)
	safeDiff := NormalizeDifficulty(req.Difficulty)

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate %d unique multiple-choice question(s) for the following curriculum topic:\n\n", req.Count))
	sb.WriteString(fmt.Sprintf("Topic: %s (%s, %s)\n", topic.Title, topic.ModuleTitle, topic.WeekCode))
	sb.WriteString(fmt.Sprintf("Requested Question Style: %s\n", safeStyle))
	sb.WriteString(fmt.Sprintf("Requested Difficulty: %s\n\n", safeDiff))

	sb.WriteString("Style Instructions:\n")
	switch safeStyle {
	case models.StyleDefinition:
		sb.WriteString("- Focus on foundational concepts, core terminology, protocols, and architectural models defined in the notes.\n")
	case models.StyleCommandSyntax:
		sb.WriteString("- Focus on exact CLI commands, arguments, option flags, piping, environment variables, or script behavior (Linux/Bash/Git/AWS CLI).\n")
	case models.StyleComparison:
		sb.WriteString("- Contrast two complementary or competing tools, services, storage classes, or architectural paradigms.\n")
	case models.StyleScenario:
		sb.WriteString("- Present a concrete workplace problem (incident response, cost explosion, permission denial, or multi-AZ scaling) requiring the candidate to pick the optimal design or fix.\n")
	}

	sb.WriteString("\nCourse Material Excerpts (STRICT GROUNDING SOURCE):\n")
	sb.WriteString("====================================================\n")
	for i, exc := range excerpts {
		trimmed := strings.TrimSpace(exc)
		if trimmed != "" {
			sb.WriteString(fmt.Sprintf("[Excerpt %d]\n%s\n\n", i+1, trimmed))
		}
	}
	sb.WriteString("====================================================\n\n")

	if len(req.AvoidStems) > 0 {
		sb.WriteString("DO NOT generate questions similar to these existing question stems (avoid duplicates):\n")
		limit := 10
		if len(req.AvoidStems) < limit {
			limit = len(req.AvoidStems)
		}
		for _, stem := range req.AvoidStems[:limit] {
			sb.WriteString(fmt.Sprintf("- %s\n", stem))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Return JSON with this exact schema:
{
  "questions": [
    {
      "stem": "string",
      "questionStyle": "` + safeStyle + `",
      "difficulty": "` + safeDiff + `",
      "cognitiveSkill": "string",
      "scenarioType": "string",
      "correctOptionKey": "A",
      "explanation": "string",
      "sourceReference": "` + topic.ModuleTitle + `, ` + topic.WeekCode + `",
      "options": [
        {"key": "A", "text": "Correct answer text", "distractorRationale": ""},
        {"key": "B", "text": "Plausible distractor text", "distractorRationale": "Why B is wrong"},
        {"key": "C", "text": "Plausible distractor text", "distractorRationale": "Why C is wrong"},
        {"key": "D", "text": "Plausible distractor text", "distractorRationale": "Why D is wrong"}
      ]
    }
  ]
}`)

	return sb.String()
}

