// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import "errors"

type ScanState int

const (
	TrueConditionState ScanState = iota
	FalseConditionState
	InElseState
)

var ErrNoCondition = errors.New("not within a conditional section")

type Conditional struct {
	states []ScanState
}

func (c *Conditional) IsActive() bool {
	return len(c.states) > 0
}

func (c *Conditional) PushState(state ScanState) {
	c.states = append(c.states, state)
}

func (c *Conditional) PopState() error {
	if len(c.states) == 0 {
		return ErrNoCondition
	}
	c.states = c.states[:len(c.states)-1]
	return nil
}

func (c *Conditional) SetState(state ScanState) {
	c.states[len(c.states)-1] = state
}

func (c *Conditional) CurrentState() ScanState {
	return c.states[len(c.states)-1]
}

func (c *Conditional) IsSkippingLines() bool {
	if !c.IsActive() {
		return false
	}
	if c.CurrentState() == TrueConditionState {
		return false
	}
	return true
}

func (c *Conditional) IsNested() bool {
	return len(c.states) > 1
}

func (c *Conditional) SkipLine(line string) bool {
	if c.CurrentState() == TrueConditionState {
		return false
	}
	return false
}

func (c *Conditional) ToggleState() {
	switch c.CurrentState() {
	case TrueConditionState:
		c.SetState(FalseConditionState)
	case FalseConditionState:
		c.SetState(TrueConditionState)
	}
}
