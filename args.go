package main

import (
	"fmt"
	"strings"
)

// Command constants define supported resources, actions, and flags.
const (
	resourceWorkbook = "workbook"
	resourceSheet    = "sheet"
	resourceCell     = "cell"
	resourceRange    = "range"

	actionInfo  = "info"
	actionList  = "list"
	actionRead  = "read"
	actionSet   = "set"
	actionClear = "clear"

	flagPretty  = "--pretty"
	flagSheet   = "--sheet"
	flagCell    = "--cell"
	flagRange   = "--range"
	flagValue   = "--value"
	flagFormula = "--formula"
)

// parsedArgs holds a validated CLI command.
type parsedArgs struct {
	resource   string
	action     string
	file       string
	sheet      string
	cell       string
	cellRange  string
	value      string
	formula    string
	valueSet   bool
	formulaSet bool
	pretty     bool
}

// parseError describes a usage error from argument parsing.
type parseError struct {
	message string
	pretty  bool
}

// Error returns the parse error message.
func (e parseError) Error() string {
	return e.message
}

// commandSpec describes valid flags for a command.
type commandSpec struct {
	resource              string
	action                string
	allowSheet            bool
	allowCell             bool
	allowRange            bool
	allowValue            bool
	allowFormula          bool
	requireSheet          bool
	requireCell           bool
	requireRange          bool
	requireValueOrFormula bool
}

// argParser incrementally parses CLI arguments.
type argParser struct {
	args        []string
	spec        commandSpec
	cmd         parsedArgs
	file        string
	seenPretty  bool
	seenSheet   bool
	seenCell    bool
	seenRange   bool
	seenValue   bool
	seenFormula bool
}

// parseArgs parses and validates CLI arguments.
func parseArgs(args []string) (parsedArgs, error) {
	parser := argParser{args: args}
	return parser.parse()
}

// parse parses the configured argument slice.
func (p *argParser) parse() (parsedArgs, error) {
	spec, err := p.parseCommandSpec()
	if err != nil {
		return parsedArgs{}, err
	}

	p.spec = spec
	p.cmd = parsedArgs{
		resource: spec.resource,
		action:   spec.action,
	}

	if err := p.parseTail(); err != nil {
		return parsedArgs{}, err
	}

	p.cmd.file = p.file

	if err := p.validate(); err != nil {
		return parsedArgs{}, err
	}

	if err := p.normalize(); err != nil {
		return parsedArgs{}, err
	}

	return p.cmd, nil
}

// parseCommandSpec resolves the command resource and action.
func (p *argParser) parseCommandSpec() (commandSpec, error) {
	if len(p.args) == 0 {
		return commandSpec{}, parseError{message: "missing command"}
	}

	if len(p.args) == 1 {
		message := fmt.Sprintf("missing action for %q", p.args[0])
		return commandSpec{}, parseError{message: message}
	}

	spec, ok := lookupCommand(p.args[0], p.args[1])
	if ok {
		return spec, nil
	}

	message := fmt.Sprintf("unknown command: %s %s", p.args[0], p.args[1])
	return commandSpec{}, parseError{message: message}
}

// parseTail parses the file argument and trailing flags.
func (p *argParser) parseTail() error {
	for i := 2; i < len(p.args); i++ {
		arg := p.args[i]
		if !isFlagToken(arg) {
			if p.file != "" {
				return parseError{message: fmt.Sprintf("unexpected argument: %s", arg), pretty: p.cmd.pretty}
			}

			p.file = arg
			continue
		}

		next, err := p.parseFlag(i)
		if err != nil {
			return err
		}

		i = next
	}

	return nil
}

// parseFlag parses one supported flag at the given index.
func (p *argParser) parseFlag(index int) (int, error) {
	switch p.args[index] {
	case flagPretty:
		if p.seenPretty {
			return 0, parseError{message: fmt.Sprintf("duplicate flag: %s", flagPretty), pretty: p.cmd.pretty}
		}

		p.seenPretty = true
		p.cmd.pretty = true
		return index, nil
	case flagSheet:
		if p.seenSheet {
			return 0, parseError{message: fmt.Sprintf("duplicate flag: %s", flagSheet), pretty: p.cmd.pretty}
		}

		value, next, err := readFlagValue(p.args, index, false)
		if err != nil {
			return 0, parseError{message: err.Error(), pretty: p.cmd.pretty}
		}

		p.seenSheet = true
		p.cmd.sheet = value
		return next, nil
	case flagCell:
		if p.seenCell {
			return 0, parseError{message: fmt.Sprintf("duplicate flag: %s", flagCell), pretty: p.cmd.pretty}
		}

		value, next, err := readFlagValue(p.args, index, false)
		if err != nil {
			return 0, parseError{message: err.Error(), pretty: p.cmd.pretty}
		}

		p.seenCell = true
		p.cmd.cell = value
		return next, nil
	case flagRange:
		if p.seenRange {
			return 0, parseError{message: fmt.Sprintf("duplicate flag: %s", flagRange), pretty: p.cmd.pretty}
		}

		value, next, err := readFlagValue(p.args, index, false)
		if err != nil {
			return 0, parseError{message: err.Error(), pretty: p.cmd.pretty}
		}

		p.seenRange = true
		p.cmd.cellRange = value
		return next, nil
	case flagValue:
		if p.seenValue {
			return 0, parseError{message: fmt.Sprintf("duplicate flag: %s", flagValue), pretty: p.cmd.pretty}
		}

		value, next, err := readFlagValue(p.args, index, true)
		if err != nil {
			return 0, parseError{message: err.Error(), pretty: p.cmd.pretty}
		}

		p.seenValue = true
		p.cmd.value = value
		p.cmd.valueSet = true
		return next, nil
	case flagFormula:
		if p.seenFormula {
			return 0, parseError{message: fmt.Sprintf("duplicate flag: %s", flagFormula), pretty: p.cmd.pretty}
		}

		value, next, err := readFlagValue(p.args, index, true)
		if err != nil {
			return 0, parseError{message: err.Error(), pretty: p.cmd.pretty}
		}

		p.seenFormula = true
		p.cmd.formula = value
		p.cmd.formulaSet = true
		return next, nil
	default:
		return 0, parseError{message: fmt.Sprintf("unknown flag: %s", p.args[index]), pretty: p.cmd.pretty}
	}
}

// validate checks parsed arguments against the command spec.
func (p *argParser) validate() error {
	if p.cmd.file == "" {
		return parseError{message: "missing file", pretty: p.cmd.pretty}
	}

	if p.seenSheet && !p.spec.allowSheet {
		return parseError{message: fmt.Sprintf("flag not allowed: %s", flagSheet), pretty: p.cmd.pretty}
	}

	if p.seenCell && !p.spec.allowCell {
		return parseError{message: fmt.Sprintf("flag not allowed: %s", flagCell), pretty: p.cmd.pretty}
	}

	if p.seenRange && !p.spec.allowRange {
		return parseError{message: fmt.Sprintf("flag not allowed: %s", flagRange), pretty: p.cmd.pretty}
	}

	if p.seenValue && !p.spec.allowValue {
		return parseError{message: fmt.Sprintf("flag not allowed: %s", flagValue), pretty: p.cmd.pretty}
	}

	if p.seenFormula && !p.spec.allowFormula {
		return parseError{message: fmt.Sprintf("flag not allowed: %s", flagFormula), pretty: p.cmd.pretty}
	}

	if p.spec.requireSheet && p.cmd.sheet == "" {
		return parseError{message: fmt.Sprintf("missing %s", flagSheet), pretty: p.cmd.pretty}
	}

	if p.spec.requireCell && p.cmd.cell == "" {
		return parseError{message: fmt.Sprintf("missing %s", flagCell), pretty: p.cmd.pretty}
	}

	if p.spec.requireRange && p.cmd.cellRange == "" {
		return parseError{message: fmt.Sprintf("missing %s", flagRange), pretty: p.cmd.pretty}
	}

	if p.spec.requireValueOrFormula {
		if p.cmd.valueSet && p.cmd.formulaSet {
			return parseError{message: fmt.Sprintf("cannot combine %s and %s", flagValue, flagFormula), pretty: p.cmd.pretty}
		}

		if !p.cmd.valueSet && !p.cmd.formulaSet {
			return parseError{message: fmt.Sprintf("missing %s or %s", flagValue, flagFormula), pretty: p.cmd.pretty}
		}
	}

	return nil
}

// normalize canonicalizes parsed cell and range references.
func (p *argParser) normalize() error {
	if p.cmd.cell != "" {
		normalized, err := normalizeCellRef(p.cmd.cell)
		if err != nil {
			return parseError{message: err.Error(), pretty: p.cmd.pretty}
		}

		p.cmd.cell = normalized
	}

	if p.cmd.cellRange != "" {
		normalized, err := parseCellRange(p.cmd.cellRange)
		if err != nil {
			return parseError{message: err.Error(), pretty: p.cmd.pretty}
		}

		p.cmd.cellRange = normalized.ref
	}

	if p.cmd.formulaSet {
		p.cmd.formula = strings.TrimPrefix(p.cmd.formula, "=")
		if p.cmd.formula == "" {
			return parseError{message: fmt.Sprintf("empty %s", flagFormula), pretty: p.cmd.pretty}
		}
	}

	return nil
}

// lookupCommand returns the spec for a supported command.
func lookupCommand(resource, action string) (commandSpec, bool) {
	switch {
	case resource == resourceWorkbook && action == actionInfo:
		return commandSpec{resource: resource, action: action}, true
	case resource == resourceSheet && action == actionList:
		return commandSpec{resource: resource, action: action}, true
	case resource == resourceSheet && action == actionInfo:
		return commandSpec{
			resource:     resource,
			action:       action,
			allowSheet:   true,
			requireSheet: true,
		}, true
	case resource == resourceCell && action == actionRead:
		return commandSpec{
			resource:     resource,
			action:       action,
			allowSheet:   true,
			allowCell:    true,
			requireSheet: true,
			requireCell:  true,
		}, true
	case resource == resourceCell && action == actionSet:
		return commandSpec{
			resource:              resource,
			action:                action,
			allowSheet:            true,
			allowCell:             true,
			allowValue:            true,
			allowFormula:          true,
			requireCell:           true,
			requireValueOrFormula: true,
		}, true
	case resource == resourceCell && action == actionClear:
		return commandSpec{
			resource:    resource,
			action:      action,
			allowSheet:  true,
			allowCell:   true,
			requireCell: true,
		}, true
	case resource == resourceRange && action == actionRead:
		return commandSpec{
			resource:     resource,
			action:       action,
			allowSheet:   true,
			allowRange:   true,
			requireSheet: true,
			requireRange: true,
		}, true
	case resource == resourceRange && action == actionClear:
		return commandSpec{
			resource:     resource,
			action:       action,
			allowSheet:   true,
			allowRange:   true,
			requireRange: true,
		}, true
	default:
		return commandSpec{}, false
	}
}

// readFlagValue reads the value following a flag token.
func readFlagValue(args []string, index int, allowFlagTokenValue bool) (string, int, error) {
	next := index + 1
	if next >= len(args) || (!allowFlagTokenValue && isFlagToken(args[next])) {
		return "", 0, fmt.Errorf("missing value for %s", args[index])
	}

	return args[next], next, nil
}

// isFlagToken reports whether an argument looks like a flag.
func isFlagToken(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}
