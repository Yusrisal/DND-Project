package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

// Dice expressions look like 1d20+5, 2d6, 4d6kh3 or 2d20kl1-1: a sum of terms,
// where a term is either a flat number or NdM with an optional keep-highest or
// keep-lowest suffix. That covers advantage (2d20kh1), disadvantage (2d20kl1)
// and ability generation (4d6kh3) without inventing a language.
const (
	maxDiceTerms = 20
	maxDiceCount = 100
	maxDieSides  = 1000
)

// DiceGroup is one NdM term with every die shown, so a roll can be checked
// rather than taken on faith.
type DiceGroup struct {
	Expression string `json:"expression"`
	Rolls      []int  `json:"rolls"`
	Dropped    []int  `json:"dropped"`
	Total      int    `json:"total"`
}

type RollResult struct {
	Expression string      `json:"expression"`
	Groups     []DiceGroup `json:"groups"`
	Modifier   int         `json:"modifier"`
	Total      int         `json:"total"`
}

// EvaluateRoll parses and rolls an expression. It is the only place dice are
// produced: the client never sends a result, only the expression.
func EvaluateRoll(expression string) (RollResult, error) {
	cleaned := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(expression)), " ", "")
	if cleaned == "" {
		return RollResult{}, invalid("expression", "is required")
	}

	terms, err := splitTerms(cleaned)
	if err != nil {
		return RollResult{}, err
	}

	result := RollResult{Expression: cleaned}
	for _, term := range terms {
		if !strings.Contains(term.body, "d") {
			value, err := strconv.Atoi(term.body)
			if err != nil {
				return RollResult{}, invalid("expression", fmt.Sprintf("cannot read %q", term.body))
			}
			result.Modifier += term.sign * value
			continue
		}

		group, err := rollGroup(term.body)
		if err != nil {
			return RollResult{}, err
		}
		if term.sign < 0 {
			group.Expression = "-" + group.Expression
			group.Total = -group.Total
		}
		result.Groups = append(result.Groups, group)
	}

	result.Total = result.Modifier
	for _, group := range result.Groups {
		result.Total += group.Total
	}
	return result, nil
}

type diceTerm struct {
	sign int
	body string
}

func splitTerms(expression string) ([]diceTerm, error) {
	var terms []diceTerm
	sign := 1
	current := strings.Builder{}

	flush := func() error {
		if current.Len() == 0 {
			return invalid("expression", "has an empty term")
		}
		terms = append(terms, diceTerm{sign: sign, body: current.String()})
		current.Reset()
		return nil
	}

	for i, char := range expression {
		switch {
		case char == '+' || char == '-':
			if i == 0 {
				if char == '-' {
					sign = -1
				}
				continue
			}
			if err := flush(); err != nil {
				return nil, err
			}
			sign = 1
			if char == '-' {
				sign = -1
			}
		default:
			current.WriteRune(char)
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}

	if len(terms) > maxDiceTerms {
		return nil, invalid("expression", "has too many terms")
	}
	return terms, nil
}

func rollGroup(body string) (DiceGroup, error) {
	keepHighest, keepLowest := 0, 0
	dicePart := body

	if index := strings.Index(body, "kh"); index >= 0 {
		dicePart = body[:index]
		n, err := strconv.Atoi(body[index+2:])
		if err != nil || n <= 0 {
			return DiceGroup{}, invalid("expression", "kh needs a positive number")
		}
		keepHighest = n
	} else if index := strings.Index(body, "kl"); index >= 0 {
		dicePart = body[:index]
		n, err := strconv.Atoi(body[index+2:])
		if err != nil || n <= 0 {
			return DiceGroup{}, invalid("expression", "kl needs a positive number")
		}
		keepLowest = n
	}

	countText, sidesText, found := strings.Cut(dicePart, "d")
	if !found {
		return DiceGroup{}, invalid("expression", fmt.Sprintf("cannot read %q", body))
	}

	count := 1
	if countText != "" {
		parsed, err := strconv.Atoi(countText)
		if err != nil {
			return DiceGroup{}, invalid("expression", fmt.Sprintf("cannot read %q", body))
		}
		count = parsed
	}
	sides, err := strconv.Atoi(sidesText)
	if err != nil {
		return DiceGroup{}, invalid("expression", fmt.Sprintf("cannot read %q", body))
	}

	if count < 1 || count > maxDiceCount {
		return DiceGroup{}, invalid("expression", fmt.Sprintf("dice count must be between 1 and %d", maxDiceCount))
	}
	if sides < 2 || sides > maxDieSides {
		return DiceGroup{}, invalid("expression", fmt.Sprintf("die must have between 2 and %d sides", maxDieSides))
	}
	if keepHighest > count || keepLowest > count {
		return DiceGroup{}, invalid("expression", "cannot keep more dice than are rolled")
	}

	rolls := make([]int, count)
	for i := range rolls {
		value, err := rollDie(sides)
		if err != nil {
			return DiceGroup{}, err
		}
		rolls[i] = value
	}

	group := DiceGroup{Expression: body, Rolls: rolls, Dropped: []int{}}
	kept := append([]int(nil), rolls...)

	if keepHighest > 0 || keepLowest > 0 {
		sort.Ints(kept)
		if keepHighest > 0 {
			group.Dropped = append(group.Dropped, kept[:count-keepHighest]...)
			kept = kept[count-keepHighest:]
		} else {
			group.Dropped = append(group.Dropped, kept[keepLowest:]...)
			kept = kept[:keepLowest]
		}
	}

	for _, value := range kept {
		group.Total += value
	}
	return group, nil
}

// rollDie uses crypto/rand so nobody has to wonder whether the server seeded
// its generator predictably.
func rollDie(sides int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 1, nil
}
