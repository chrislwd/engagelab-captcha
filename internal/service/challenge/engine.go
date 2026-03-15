package challenge

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/engagelab/captcha/internal/model"
	"github.com/google/uuid"
)

// Engine generates and validates CAPTCHA challenges.
type Engine struct {
	mu sync.RWMutex

	// Stores the expected answer for each challenge ID.
	answers   map[string]answerRecord
	hmacKey   []byte
}

type answerRecord struct {
	ChallengeType model.ChallengeType
	Data          interface{} // type-specific expected answer
	ExpiresAt     time.Time
}

// sliderAnswer holds the expected X position for a slider challenge.
type sliderAnswer struct {
	TargetX  int `json:"target_x"`
	Tolerance int `json:"tolerance"`
}

// clickAnswer holds the expected click targets.
type clickAnswer struct {
	TargetIDs []string `json:"target_ids"`
}

// NewEngine creates a new challenge engine.
func NewEngine(hmacSecret string) *Engine {
	return &Engine{
		answers: make(map[string]answerRecord),
		hmacKey: []byte(hmacSecret),
	}
}

// GenerateChallenge produces a challenge configuration for the given type.
func (e *Engine) GenerateChallenge(challengeType model.ChallengeType) model.ChallengeConfig {
	challengeID := uuid.NewString()
	expiresAt := time.Now().Add(5 * time.Minute)

	var config model.ChallengeConfig

	switch challengeType {
	case model.ChallengeTypeSlider:
		config = e.generateSlider(challengeID, expiresAt)
	case model.ChallengeTypeClick:
		config = e.generateClick(challengeID, expiresAt)
	case model.ChallengeTypePuzzle:
		config = e.generatePuzzle(challengeID, expiresAt)
	case model.ChallengeTypeInvisible:
		config = e.generateInvisible(challengeID, expiresAt)
	default:
		config = e.generateSlider(challengeID, expiresAt)
	}

	return config
}

func (e *Engine) generateSlider(challengeID string, expiresAt time.Time) model.ChallengeConfig {
	bgWidth := 320
	bgHeight := 160
	sliderW := 50
	sliderH := 50

	// Random target X position within the valid range.
	targetX := 60 + rand.Intn(bgWidth-sliderW-80)
	targetY := 20 + rand.Intn(bgHeight-sliderH-40)

	e.mu.Lock()
	e.answers[challengeID] = answerRecord{
		ChallengeType: model.ChallengeTypeSlider,
		Data:          sliderAnswer{TargetX: targetX, Tolerance: 5},
		ExpiresAt:     expiresAt,
	}
	e.mu.Unlock()

	return model.ChallengeConfig{
		ChallengeID:    challengeID,
		ChallengeType:  model.ChallengeTypeSlider,
		SliderBgWidth:  bgWidth,
		SliderBgHeight: bgHeight,
		SliderX:        0, // starting position sent to client
		SliderY:        targetY,
		SliderWidth:    sliderW,
		SliderHeight:   sliderH,
		ExpiresAt:      expiresAt,
	}
}

func (e *Engine) generateClick(challengeID string, expiresAt time.Time) model.ChallengeConfig {
	// Generate a set of targets; a subset are the "correct" ones.
	allLabels := []string{"cat", "dog", "car", "tree", "house", "bird", "fish", "star", "moon", "sun"}
	rand.Shuffle(len(allLabels), func(i, j int) { allLabels[i], allLabels[j] = allLabels[j], allLabels[i] })

	numTargets := 6
	numCorrect := 3
	if numTargets > len(allLabels) {
		numTargets = len(allLabels)
	}

	promptLabel := allLabels[0] // the category to select
	targets := make([]model.ClickTarget, numTargets)
	correctIDs := make([]string, 0, numCorrect)

	for i := 0; i < numTargets; i++ {
		tid := uuid.NewString()[:8]
		targets[i] = model.ClickTarget{
			ID:    tid,
			Label: allLabels[i%len(allLabels)],
			X:     30 + rand.Intn(260),
			Y:     20 + rand.Intn(130),
		}
		// First numCorrect targets that match the prompt are correct.
		if targets[i].Label == promptLabel && len(correctIDs) < numCorrect {
			correctIDs = append(correctIDs, tid)
		}
	}

	// Ensure at least the first target matches the prompt for a solvable challenge.
	if len(correctIDs) == 0 {
		targets[0].Label = promptLabel
		correctIDs = append(correctIDs, targets[0].ID)
	}

	e.mu.Lock()
	e.answers[challengeID] = answerRecord{
		ChallengeType: model.ChallengeTypeClick,
		Data:          clickAnswer{TargetIDs: correctIDs},
		ExpiresAt:     expiresAt,
	}
	e.mu.Unlock()

	return model.ChallengeConfig{
		ChallengeID:   challengeID,
		ChallengeType: model.ChallengeTypeClick,
		ClickTargets:  targets,
		ClickPrompt:   fmt.Sprintf("Select all images of: %s", promptLabel),
		ExpiresAt:     expiresAt,
	}
}

func (e *Engine) generatePuzzle(challengeID string, expiresAt time.Time) model.ChallengeConfig {
	// Puzzle: user must drag a piece to the correct X position (similar to slider but with image).
	targetX := 50 + rand.Intn(220)
	targetY := 30 + rand.Intn(100)

	e.mu.Lock()
	e.answers[challengeID] = answerRecord{
		ChallengeType: model.ChallengeTypePuzzle,
		Data:          sliderAnswer{TargetX: targetX, Tolerance: 4},
		ExpiresAt:     expiresAt,
	}
	e.mu.Unlock()

	return model.ChallengeConfig{
		ChallengeID:    challengeID,
		ChallengeType:  model.ChallengeTypePuzzle,
		PuzzleImageURL: fmt.Sprintf("/assets/puzzles/%d.png", rand.Intn(20)+1),
		PuzzlePieceX:   targetX,
		PuzzlePieceY:   targetY,
		ExpiresAt:      expiresAt,
	}
}

func (e *Engine) generateInvisible(challengeID string, expiresAt time.Time) model.ChallengeConfig {
	// Invisible: auto-pass, generate a signed token immediately.
	token := e.signToken(challengeID)

	e.mu.Lock()
	e.answers[challengeID] = answerRecord{
		ChallengeType: model.ChallengeTypeInvisible,
		Data:          nil,
		ExpiresAt:     expiresAt,
	}
	e.mu.Unlock()

	return model.ChallengeConfig{
		ChallengeID:   challengeID,
		ChallengeType: model.ChallengeTypeInvisible,
		Token:         token,
		ExpiresAt:     expiresAt,
	}
}

// ValidateChallenge checks the user's answer against the expected answer.
// Returns true if valid, along with a signed token for server-side verification.
func (e *Engine) ValidateChallenge(challengeID string, answer interface{}) (bool, string) {
	e.mu.Lock()
	rec, ok := e.answers[challengeID]
	if !ok {
		e.mu.Unlock()
		return false, ""
	}
	// Prevent replay: delete the answer once validated.
	delete(e.answers, challengeID)
	e.mu.Unlock()

	if time.Now().After(rec.ExpiresAt) {
		return false, ""
	}

	switch rec.ChallengeType {
	case model.ChallengeTypeSlider, model.ChallengeTypePuzzle:
		return e.validateSliderAnswer(rec, answer, challengeID)
	case model.ChallengeTypeClick:
		return e.validateClickAnswer(rec, answer, challengeID)
	case model.ChallengeTypeInvisible:
		// Invisible challenges are always valid.
		return true, e.signToken(challengeID)
	}

	return false, ""
}

func (e *Engine) validateSliderAnswer(rec answerRecord, answer interface{}, challengeID string) (bool, string) {
	expected, ok := rec.Data.(sliderAnswer)
	if !ok {
		return false, ""
	}

	var submittedX float64
	switch v := answer.(type) {
	case float64:
		submittedX = v
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return false, ""
		}
		submittedX = f
	case map[string]interface{}:
		if x, ok := v["x"]; ok {
			switch xv := x.(type) {
			case float64:
				submittedX = xv
			default:
				return false, ""
			}
		}
	default:
		return false, ""
	}

	diff := math.Abs(submittedX - float64(expected.TargetX))
	if diff <= float64(expected.Tolerance) {
		return true, e.signToken(challengeID)
	}
	return false, ""
}

func (e *Engine) validateClickAnswer(rec answerRecord, answer interface{}, challengeID string) (bool, string) {
	expected, ok := rec.Data.(clickAnswer)
	if !ok {
		return false, ""
	}

	// Answer should be a list of target IDs.
	var submittedIDs []string
	switch v := answer.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				submittedIDs = append(submittedIDs, s)
			}
		}
	case []string:
		submittedIDs = v
	default:
		return false, ""
	}

	if len(submittedIDs) != len(expected.TargetIDs) {
		return false, ""
	}

	expectedSet := make(map[string]bool)
	for _, id := range expected.TargetIDs {
		expectedSet[id] = true
	}
	for _, id := range submittedIDs {
		if !expectedSet[id] {
			return false, ""
		}
	}

	return true, e.signToken(challengeID)
}

// signToken creates an HMAC-signed token that encodes the challenge ID and timestamp.
func (e *Engine) signToken(challengeID string) string {
	ts := time.Now().UTC().Format(time.RFC3339)
	payload := fmt.Sprintf("%s|%s", challengeID, ts)
	mac := hmac.New(sha256.New, e.hmacKey)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", payload, sig)
}

// CleanExpired removes expired answer records. Should be called periodically.
func (e *Engine) CleanExpired() {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now()
	for id, rec := range e.answers {
		if now.After(rec.ExpiresAt) {
			delete(e.answers, id)
		}
	}
}
