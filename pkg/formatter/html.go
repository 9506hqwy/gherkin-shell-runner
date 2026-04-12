package formatter

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
)

func init() {
	godog.Format("html", "HTML formatter", htmlFormatterFunc)
}

func htmlFormatterFunc(suite string, out io.Writer) godog.Formatter {
	return newHTMLFmt(suite, out)
}

func newHTMLFmt(suite string, out io.Writer) *htmlFmt {
	return &htmlFmt{
		BaseFmt: godog.NewBaseFmt(suite, out),
		out:     out,
	}
}

// ---------------------------------------------------------------------------

type htmlFmt struct {
	*godog.BaseFmt
	out io.Writer
}

func (f *htmlFmt) Summary() {
	features := f.Storage.MustGetFeatures()

	var files []*fileProc
	for _, feature := range features {
		file := fileProc{
			URI: feature.Uri,
		}
		files = append(files, &file)

		comments := collectProcInComments(feature.Comments)

		f.collectProcInFeature(
			feature.FindScenario,
			feature.FindExample,
			feature.FindStep,
			comments,
			feature.Feature,
			&file,
		)

		complementLine(&file)
	}

	err := marshalHTML(files, f.out)
	if err != nil {
		panic(err)
	}
}

func (f *htmlFmt) collectProcInFeature(
	findScenario func(id string) *messages.Scenario,
	findExample func(id string) (*messages.Examples, *messages.TableRow),
	findStep func(id string) *messages.Step,
	comments []procInterface,
	feature *messages.Feature,
	file *fileProc,
) *fileProc {
	var procs []procInterface

	if feature == nil {
		return file
	}

	procs = append(procs, createFeatureLine(feature))

	tagLines := collectProcInTags(feature.Tags)
	procs = append(procs, tagLines...)

	pickles := f.Storage.MustGetPickles(file.URI)
	for _, scenarioPickle := range pickles {
		scenarioLines := f.collectProcInScenarioExec(
			findScenario,
			findExample,
			findStep,
			comments,
			feature.Children,
			scenarioPickle,
		)
		procs = append(procs, scenarioLines...)
	}

	procs = insertComments(procs, comments)

	lines := collectLines(procs)
	file.Lines = append(file.Lines, lines...)

	return file
}

func (f *htmlFmt) collectProcInScenarioExec(
	findScenario func(id string) *messages.Scenario,
	findExample func(id string) (*messages.Examples, *messages.TableRow),
	findStep func(id string) *messages.Step,
	comments []procInterface,
	children []*messages.FeatureChild,
	scenarioPickle *messages.Pickle,
) []procInterface {
	var procs []procInterface

	scenario := findScenario(scenarioPickle.AstNodeIds[0])

	scenarioResult := f.Storage.MustGetPickleResult(scenarioPickle.Id)
	startedAt := &scenarioResult.StartedAt

	stepResults := f.collectProcInSteps(findStep, scenarioPickle, startedAt)

	for _, child := range children {
		if child.Scenario != nil && scenario.Id != child.Scenario.Id {
			continue
		}

		if hasScenarioInRule(child.Rule, scenario) {
			ruleLines := collectProcInRule(
				findExample,
				child.Rule,
				stepResults,
				comments,
				scenarioPickle,
				scenario,
				startedAt,
			)
			procs = append(procs, ruleLines...)
		}

		bgLines := collectProcInBackground(
			child.Background,
			stepResults,
			comments,
		)
		procs = append(procs, bgLines...)

		scenarioLine := collectProcInScenario(
			findExample,
			stepResults,
			comments,
			scenarioPickle,
			child.Scenario,
			startedAt,
		)
		procs = append(procs, scenarioLine...)
	}

	block := &blockProc{
		lines: procs,
	}

	return []procInterface{block}
}

func (f *htmlFmt) collectProcInSteps(
	findStep func(id string) *messages.Step,
	scenarioPickle *messages.Pickle,
	startedAt *time.Time,
) map[int64][]procInterface {
	procs := make(map[int64][]procInterface)

	results := f.Storage.MustGetPickleStepResultsByPickleID(
		scenarioPickle.Id,
	)

	for _, result := range results {
		outputs := map[string]string{}
		for _, att := range result.Attachments {
			if att.MimeType == "text/plain" {
				outputs[att.Name] = string(att.Data)
			}
		}

		step, stepResult := createStepResult(
			f.Storage.MustGetPickleStep,
			findStep,
			result.PickleStepID,
			startedAt,
			&result.FinishedAt,
			int(result.Status),
			result.Err,
			outputs,
		)

		procs[step.Location.Line] = collectProcInStep(step, stepResult)
	}

	return procs
}

// ---------------------------------------------------------------------------

func insertComments(
	procs []procInterface,
	comments []procInterface,
) []procInterface {
	lines := collectLines(procs)

	//revive:disable
	sort.Sort(sortLinesByLocation(procs))
	//revive:enable

	b := procs[0].GetLocation()
	e := procs[len(procs)-1].GetLocation()

	for _, comment := range comments {
		loc := comment.GetLocation()
		if loc.Line >= b.Line && loc.Line <= e.Line {
			if !hasComment(lines, loc) {
				procs = append(procs, comment)
			}
		}
	}

	//revive:disable
	sort.Sort(sortLinesByLocation(procs))
	//revive:enable

	return procs
}

func hasComment(lines []*lineProc, loc messages.Location) bool {
	for _, l := range lines {
		if l.Location.Line == loc.Line {
			return true
		}
	}

	return false
}

func hasScenarioInRule(rule *messages.Rule, scenrio *messages.Scenario) bool {
	if rule == nil {
		return false
	}

	for _, child := range rule.Children {
		if child.Scenario != nil && child.Scenario.Id == scenrio.Id {
			return true
		}
	}

	return false
}

func collectLines(procs []procInterface) []*lineProc {
	var lines []*lineProc

	for _, proc := range procs {
		l := collectLine(proc)
		lines = append(lines, l...)
	}

	return lines
}

func collectLine(proc procInterface) []*lineProc {
	if line, ok := proc.(*lineProc); ok {
		return []*lineProc{line}
	}

	if block, ok := proc.(*blockProc); ok {
		return collectLines(block.lines)
	}

	return []*lineProc{}
}

func complementLine(file *fileProc) {
	for _, line := range file.Lines {
		line.file = file
	}
}

func collectProcInComments(comments []*messages.Comment) []procInterface {
	var procs []procInterface

	for _, comment := range comments {
		commentLines := collectProcInComment(comment)
		procs = append(procs, commentLines...)
	}

	return procs
}

func collectProcInComment(comment *messages.Comment) []procInterface {
	var procs []procInterface

	if comment == nil {
		return procs
	}

	procs = append(procs, createCommentLine(comment))

	return procs
}

func collectProcInTags(tags []*messages.Tag) []procInterface {
	var procs []procInterface

	for _, tag := range tags {
		tagLines := collectProcInTag(tag)
		procs = append(procs, tagLines...)
	}

	return procs
}

func collectProcInTag(tag *messages.Tag) []procInterface {
	var procs []procInterface

	if tag == nil {
		return procs
	}

	procs = append(procs, createTagLine(tag))

	return procs
}

func collectProcInRule(
	findExample func(id string) (*messages.Examples, *messages.TableRow),
	rule *messages.Rule,
	results map[int64][]procInterface,
	comments []procInterface,
	scenarioPickle *messages.Pickle,
	scenario *messages.Scenario,
	startedAt *time.Time,
) []procInterface {
	var procs []procInterface

	if rule == nil {
		return procs
	}

	procs = append(procs, createRuleLine(rule))

	tagLines := collectProcInTags(rule.Tags)
	procs = append(procs, tagLines...)

	for _, child := range rule.Children {
		if child.Scenario != nil && child.Scenario.Id != scenario.Id {
			continue
		}

		bgLines := collectProcInBackground(child.Background, results, comments)
		procs = append(procs, bgLines...)

		scenarioLine := collectProcInScenario(
			findExample,
			results,
			comments,
			scenarioPickle,
			child.Scenario,
			startedAt)
		procs = append(procs, scenarioLine...)
	}

	return procs
}

func collectProcInBackground(
	bg *messages.Background,
	results map[int64][]procInterface,
	comments []procInterface,
) []procInterface {
	var procs []procInterface

	if bg == nil {
		return procs
	}

	procs = append(procs, createBackgroundLine(bg))

	for _, step := range bg.Steps {
		stepLines := results[step.Location.Line]
		procs = append(procs, stepLines...)
	}

	procs = insertComments(procs, comments)

	return procs
}

func collectProcInScenario(
	findExample func(id string) (*messages.Examples, *messages.TableRow),
	results map[int64][]procInterface,
	comments []procInterface,
	scenarioPickle *messages.Pickle,
	scenario *messages.Scenario,
	startedAt *time.Time,
) []procInterface {
	var procs []procInterface

	if scenario == nil {
		return procs
	}

	scenarioLine := createScenarioLine(scenario, startedAt)
	procs = append(procs, scenarioLine)

	tagLines := collectProcInTags(scenario.Tags)
	procs = append(procs, tagLines...)

	for _, step := range scenario.Steps {
		stepLines := results[step.Location.Line]
		procs = append(procs, stepLines...)
	}

	var exampleLocation *messages.Location
	//revive:disable:add-constant
	if len(scenarioPickle.AstNodeIds) > 1 {
		_, row := findExample(scenarioPickle.AstNodeIds[1])
		exampleLocation = row.Location
	}
	//revive:enable:add-constant

	exampleLines := collectProcInExamples(scenario.Examples, exampleLocation)
	procs = append(procs, exampleLines...)

	procs = insertComments(procs, comments)

	return procs
}

func collectProcInStep(
	step *messages.Step,
	result *stepResult,
) []procInterface {
	var procs []procInterface

	stepLine := createStepLine(step, result)
	procs = append(procs, stepLine)

	followResult := &stepResult{
		StartedAt:  result.StartedAt,
		FinishedAt: result.FinishedAt,
		Duration:   result.Duration,
		Status:     result.Status,
	}

	docStringProc := collectProcInDocString(step.DocString, followResult)
	procs = append(procs, docStringProc...)

	if step.DataTable != nil {
		tableLines := collectProcInTableRows(
			step.DataTable.Rows,
			LineTypeTableRow,
			followResult,
		)
		procs = append(procs, tableLines...)
	}

	return procs
}

func collectProcInExamples(
	examples []*messages.Examples,
	current *messages.Location,
) []procInterface {
	var procs []procInterface

	for _, example := range examples {
		egLines := collectProcInExample(example, current)
		procs = append(procs, egLines...)
	}

	return procs
}

func collectProcInExample(
	examples *messages.Examples,
	current *messages.Location,
) []procInterface {
	var procs []procInterface

	if examples == nil {
		return procs
	}

	procs = append(procs, createExamplesLine(examples))

	header := collectProcInTableRow(
		examples.TableHeader,
		LineTypeTableHeader,
		nil,
	)
	procs = append(procs, header...)

	for _, row := range examples.TableBody {
		var result *stepResult
		if row.Location.Line == current.Line {
			result = &stepResult{
				Status: LineStatusPassed,
			}
		}

		r := collectProcInTableRow(row, LineTypeTableRow, result)
		procs = append(procs, r...)
	}

	return procs
}

func collectProcInDocString(
	doc *messages.DocString,
	result *stepResult,
) []procInterface {
	var procs []procInterface

	if doc == nil {
		return procs
	}

	// before """
	before := *doc.Location
	procs = append(procs, createDocStringLine(before, "\"\"\"", result))

	// data
	loc := before
	for i, line := range strings.Split(doc.Content, "\n") {
		loc = before
		loc.Line += int64(i + 1)
		procs = append(procs, createDocStringLine(loc, line, result))
	}

	// after """
	after := loc
	after.Line++
	procs = append(procs, createDocStringLine(after, "\"\"\"", result))

	return procs
}

func collectProcInTableRows(
	rows []*messages.TableRow,
	lineType LineType,
	result *stepResult,
) []procInterface {
	var procs []procInterface

	for _, row := range rows {
		r := collectProcInTableRow(row, lineType, result)
		procs = append(procs, r...)
	}

	return procs
}

func collectProcInTableRow(
	row *messages.TableRow,
	lineType LineType,
	result *stepResult,
) []procInterface {
	var procs []procInterface

	if row == nil {
		return procs
	}

	values := []string{}
	for _, c := range row.Cells {
		values = append(values, c.Value)
	}

	procs = append(procs, createTableRowLine(row, lineType, values, result))

	return procs
}

func createFeatureLine(
	feature *messages.Feature,
) *lineProc {
	return &lineProc{
		LineType:    LineTypeFeature,
		Location:    *feature.Location,
		Keyword:     feature.Keyword,
		Name:        feature.Name,
		Description: feature.Description,
	}
}

func createCommentLine(comment *messages.Comment) *lineProc {
	return &lineProc{
		LineType: LineTypeComment,
		Location: *comment.Location,
		Text:     comment.Text,
	}
}

func createRuleLine(rule *messages.Rule) *lineProc {
	return &lineProc{
		LineType:    LineTypeRule,
		ID:          rule.Id,
		Location:    *rule.Location,
		Keyword:     rule.Keyword,
		Name:        rule.Name,
		Description: rule.Description,
	}
}

func createBackgroundLine(bg *messages.Background) *lineProc {
	return &lineProc{
		LineType:    LineTypeBackground,
		ID:          bg.Id,
		Location:    *bg.Location,
		Keyword:     bg.Keyword,
		Name:        bg.Name,
		Description: bg.Description,
	}
}

func createTagLine(tag *messages.Tag) *lineProc {
	return &lineProc{
		LineType: LineTypeTag,
		ID:       tag.Id,
		Location: *tag.Location,
		Text:     tag.Name,
	}
}

func createScenarioLine(
	scenario *messages.Scenario,
	startedAt *time.Time,
) *lineProc {
	return &lineProc{
		LineType:    LineTypeScenario,
		ID:          scenario.Id,
		Location:    *scenario.Location,
		Tags:        scenario.Tags,
		Keyword:     scenario.Keyword,
		Name:        scenario.Name,
		Description: scenario.Description,
		StepResult: &stepResult{
			StartedAt: startedAt,
		},
	}
}

func createStepLine(
	step *messages.Step,
	result *stepResult,
) *lineProc {
	return &lineProc{
		LineType:   LineTypeStep,
		ID:         step.Id,
		Location:   *step.Location,
		Keyword:    step.Keyword,
		Text:       step.Text,
		StepResult: result,
	}
}

func createDocStringLine(
	loc messages.Location,
	text string,
	result *stepResult,
) *lineProc {
	return &lineProc{
		LineType:   LineTypeDocString,
		Location:   loc,
		Text:       text,
		StepResult: result,
	}
}

func createExamplesLine(examples *messages.Examples) *lineProc {
	return &lineProc{
		LineType: LineTypeExamples,
		ID:       examples.Id,
		Location: *examples.Location,
		Keyword:  examples.Keyword,
		Name:     examples.Name,
	}
}

func createTableRowLine(
	row *messages.TableRow,
	lineType LineType,
	values []string,
	result *stepResult,
) *lineProc {
	return &lineProc{
		LineType:   lineType,
		Location:   *row.Location,
		Text:       fmt.Sprintf("| %s |", strings.Join(values, " | ")),
		StepResult: result,
	}
}

func createStepResult(
	getPicke func(id string) *messages.PickleStep,
	findStep func(id string) *messages.Step,
	picleID string,
	startedAt *time.Time,
	finishedAt *time.Time,
	status int,
	err error,
	outputs map[string]string,
) (*messages.Step, *stepResult) {
	pickleStep := getPicke(picleID)
	step := findStep(pickleStep.AstNodeIds[0])

	duration := finishedAt.Sub(*startedAt).Milliseconds()

	result := &stepResult{
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Duration:   duration,
		Status:     status,
		Err:        err,
		Outputs:    outputs,
	}

	return step, result
}

// ---------------------------------------------------------------------------

type procInterface interface {
	GetLocation() messages.Location
}

type sortLinesByLocation []procInterface

func (s sortLinesByLocation) Len() int {
	return len(s)
}

func (s sortLinesByLocation) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortLinesByLocation) Less(i, j int) bool {
	il := s[i].GetLocation()
	jl := s[j].GetLocation()
	return il.Line < jl.Line
}

// ---------------------------------------------------------------------------

type LineType string

const (
	LineTypeComment     LineType = "comment"
	LineTypeTag         LineType = "tag"
	LineTypeFeature     LineType = "feature"
	LineTypeRule        LineType = "rule"
	LineTypeBackground  LineType = "background"
	LineTypeScenario    LineType = "scenario"
	LineTypeStep        LineType = "step"
	LineTypeDocString   LineType = "docstring"
	LineTypeExamples    LineType = "examples"
	LineTypeTableHeader LineType = "tableheader"
	LineTypeTableRow    LineType = "tablerow"
)

const (
	LineStatusPassed    int = 0
	LineStatusFailed    int = 1
	LineStatusSkipped   int = 2
	LineStatusUndefined int = 3
	LineStatusPending   int = 4
	LineStatusAmbiguous int = 5
)

const (
	LineStatusPassedStr    string = "passed"
	LineStatusFailedStr    string = "failed"
	LineStatusSkippedStr   string = "skipped"
	LineStatusUndefinedStr string = "undefined"
	LineStatusPendingStr   string = "pending"
	LineStatusAmbiguousStr string = "ambiguous"
)

// ---------------------------------------------------------------------------

type fileProc struct {
	URI   string
	Lines []*lineProc
}

func (l *fileProc) HTMLClass() string {
	classess := []string{}

	for _, line := range l.Lines {
		if line.LineType == LineTypeStep &&
			line.StepResult != nil &&
			line.StepResult.Status == LineStatusFailed {
			classess = append(classess, LineStatusFailedStr)
			break
		}
	}

	//revive:disable:add-constant
	if len(classess) == 0 {
		classess = append(classess, LineStatusPassedStr)
	}
	//revive:enable:add-constant

	return strings.Join(classess, " ")
}

func (l *fileProc) Count(status int) int {
	//revive:disable:add-constant
	count := 0
	//revive:enable:add-constant

	for _, line := range l.Lines {
		if line.LineType != LineTypeStep {
			continue
		}

		if line.StepResult.Status != status {
			continue
		}

		count++
	}

	return count
}

// ---------------------------------------------------------------------------

type lineProc struct {
	file        *fileProc
	LineType    LineType
	ID          string
	Location    messages.Location
	Tags        []*messages.Tag
	Keyword     string
	Name        string
	Description string
	Text        string
	StepResult  *stepResult
}

func (l *lineProc) GetLocation() messages.Location {
	return l.Location
}
func (l *lineProc) HasOutputs() bool {
	//revive:disable:add-constant
	return l.StepResult != nil && len(l.StepResult.Outputs) != 0
	//revive:enable:add-constant
}

func (l *lineProc) HasSummary() bool {
	return l.LineType == LineTypeFeature || l.LineType == LineTypeScenario
}

func (l *lineProc) HTMLClass() string {
	classess := []string{string(l.LineType)}
	targetTypes := []LineType{
		LineTypeStep,
		LineTypeDocString,
		LineTypeTableHeader,
		LineTypeTableRow,
	}

	if slices.Contains(targetTypes, l.LineType) && l.StepResult != nil {
		classess = append(classess, l.StepResult.HTMLClass())
	} else {
		classess = append(classess, "unknown")
	}

	return strings.Join(classess, " ")
}

func (l *lineProc) HTMLIndent() int64 {
	//revive:disable:add-constant
	return l.Location.Column - 1
	//revive:enable:add-constant
}

func (l *lineProc) HTMLLine() string {
	text := ""

	if l.Keyword != "" {
		text += l.Keyword

		if slices.Contains([]LineType{
			LineTypeFeature,
			LineTypeRule,
			LineTypeBackground,
			LineTypeScenario,
		}, l.LineType) {
			text += ": "
		}
	}

	text += l.Name
	text += l.Text

	return text
}

//revive:disable:cognitive-complexity

func (l *lineProc) Count(status int) int {
	//revive:disable:add-constant
	count := 0
	//revive:enable:add-constant

	for _, line := range l.file.Lines {
		if line.Location.Line <= l.Location.Line {
			continue
		}

		if line.LineType == l.LineType {
			break
		}

		if line.LineType != LineTypeStep {
			continue
		}

		if line.StepResult.Status != status {
			continue
		}

		count++
	}

	return count
}

//revive:enable:cognitive-complexity

// ---------------------------------------------------------------------------

type stepResult struct {
	StartedAt  *time.Time
	FinishedAt *time.Time
	Duration   int64
	Status     int
	Err        error
	Outputs    map[string]string
}

func (s *stepResult) HTMLClass() string {
	switch s.Status {
	case LineStatusPassed:
		return "passed"
	case LineStatusFailed:
		return "failed"
	case LineStatusSkipped:
		return "skipped"
	case LineStatusUndefined:
		return "undefined"
	case LineStatusPending:
		return "pending"
	case LineStatusAmbiguous:
		return "ambiguous"
	default:
		return "unknown"
	}
}

// ---------------------------------------------------------------------------

type blockProc struct {
	lines []procInterface
}

func (b *blockProc) GetLocation() messages.Location {
	return b.lines[0].GetLocation()
}

// ---------------------------------------------------------------------------

//go:embed html.tmpl
var htmlTmpl string

type htmlTmplData struct {
	Files       []*fileProc
	PassedCount int
	FailedCount int
}

func marshalHTML(files []*fileProc, out io.Writer) error {
	t, err := template.New("gherkin").Parse(htmlTmpl)
	if err != nil {
		return err
	}

	data := htmlTmplData{
		Files: files,
	}

	for _, file := range files {
		data.PassedCount += file.Count(LineStatusPassed)
		data.FailedCount += file.Count(LineStatusFailed)
	}

	err = t.Execute(out, data)
	if err != nil {
		return err
	}

	return nil
}
