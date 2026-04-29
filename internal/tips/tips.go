// Package tips provides exercise/wellness tips shown during breaks.
// Tips are categorized so short breaks can prefer eye-focused tips and long
// breaks can rotate through stretch/move/breathe categories.
package tips

import (
	"math/rand"
	"sync"
)

// Category groups tips by the kind of activity they suggest.
type Category string

const (
	Eyes    Category = "eyes"
	Stretch Category = "stretch"
	Move    Category = "move"
	Breathe Category = "breathe"
	Posture Category = "posture"
)

// Tip is a single wellness suggestion.
type Tip struct {
	ID       int      `json:"id"`
	Text     string   `json:"text"`
	Category Category `json:"category"`
}

// Catalog is a collection of tips with random-pick utilities. Safe for
// concurrent use.
type Catalog struct {
	mu   sync.RWMutex
	tips []Tip
	rng  *rand.Rand
}

// NewCatalog returns a catalog seeded with the built-in tip set.
func NewCatalog() *Catalog {
	return &Catalog{
		tips: builtin(),
		rng:  rand.New(rand.NewSource(seed())),
	}
}

// All returns a copy of every tip.
func (c *Catalog) All() []Tip {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Tip, len(c.tips))
	copy(out, c.tips)
	return out
}

// Pick returns a random tip, optionally filtered by category. Returns a
// fallback tip if no matches exist.
func (c *Catalog) Pick(cat Category) Tip {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pool := make([]Tip, 0, len(c.tips))
	for _, t := range c.tips {
		if cat == "" || t.Category == cat {
			pool = append(pool, t)
		}
	}
	if len(pool) == 0 {
		return Tip{Text: "Take a moment to relax.", Category: Breathe}
	}
	return pool[c.rng.Intn(len(pool))]
}

// PickForShort returns a tip tailored for a short break (eye-focused).
func (c *Catalog) PickForShort() Tip { return c.Pick(Eyes) }

// PickForLong returns a tip from a randomly-chosen long-break category.
func (c *Catalog) PickForLong() Tip {
	c.mu.RLock()
	cats := []Category{Stretch, Move, Breathe}
	cat := cats[c.rng.Intn(len(cats))]
	c.mu.RUnlock()
	return c.Pick(cat)
}

// Add appends a custom tip. Returns the assigned ID.
func (c *Catalog) Add(text string, cat Category) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := 0
	for _, t := range c.tips {
		if t.ID > id {
			id = t.ID
		}
	}
	id++
	c.tips = append(c.tips, Tip{ID: id, Text: text, Category: cat})
	return id
}

// Remove deletes the tip with the given ID. Returns true if removed.
func (c *Catalog) Remove(id int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, t := range c.tips {
		if t.ID == id {
			c.tips = append(c.tips[:i], c.tips[i+1:]...)
			return true
		}
	}
	return false
}

// builtin returns the curated default tip set.
func builtin() []Tip {
	return []Tip{
		// Eyes
		{1, "Look at something 20 feet away for 20 seconds.", Eyes},
		{2, "Blink slowly 10 times to refresh your eyes.", Eyes},
		{3, "Close your eyes and roll them in slow circles.", Eyes},
		{4, "Cup your palms over closed eyes and relax.", Eyes},
		{5, "Focus on a near object, then a distant one. Repeat 5 times.", Eyes},

		// Stretch
		{10, "Stretch your arms above your head and hold for 10 seconds.", Stretch},
		{11, "Roll your shoulders backward 5 times, then forward 5 times.", Stretch},
		{12, "Tilt your head to each side, holding for 10 seconds each.", Stretch},
		{13, "Interlace your fingers and stretch your arms forward.", Stretch},
		{14, "Stand and gently twist your torso to each side.", Stretch},
		{15, "Extend each wrist and gently pull fingers back.", Stretch},

		// Move
		{20, "Stand up and walk around for a minute.", Move},
		{21, "Do 10 jumping jacks to get your blood flowing.", Move},
		{22, "March in place for 30 seconds.", Move},
		{23, "Take a short walk to refill your water.", Move},
		{24, "Do 5 slow bodyweight squats.", Move},

		// Breathe
		{30, "Inhale for 4 seconds, exhale for 6. Repeat 5 times.", Breathe},
		{31, "Box breathing: inhale 4, hold 4, exhale 4, hold 4.", Breathe},
		{32, "Close your eyes and follow your breath for 30 seconds.", Breathe},

		// Posture
		{40, "Sit tall: ears over shoulders, shoulders over hips.", Posture},
		{41, "Drop your shoulders away from your ears.", Posture},
		{42, "Adjust your screen so the top is at eye level.", Posture},
	}
}

// seed returns a deterministic-ish seed; replaced in tests.
var seed = defaultSeed
