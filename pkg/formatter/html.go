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

type htmlFmt struct {
	*godog.BaseFmt
	out io.Writer
}

func (f *htmlFmt) Summary() {
	features := f.Storage.MustGetFeatures()

	var files []*fileProc
	for _, feature := range features {
		file := collectFileInDocument(feature.GherkinDocument)
		files = append(files, file)

		pickles := f.Storage.MustGetPickles(feature.Uri)
		for _, scenarioPickle := range pickles {
			f.setResult(
				feature.FindScenario,
				feature.FindExample,
				feature.FindStep,
				scenarioPickle,
				file,
			)
		}

		complementLine(file)
	}

	err := marshalHTML(files, f.out)
	if err != nil {
		panic(err)
	}
}

func (f *htmlFmt) setResult(
	findScenario func(id string) *messages.Scenario,
	findExample func(id string) (*messages.Examples, *messages.TableRow),
	findStep func(id string) *messages.Step,
	scenarioPickle *messages.Pickle,
	file *fileProc,
) {
	scenarioLine := f.setScenarioLine(
		findScenario,
		findExample,
		scenarioPickle,
		file,
	)
	f.setStepsLine(
		findStep,
		scenarioPickle,
		file,
		scenarioLine,
	)
}

func (f *htmlFmt) setScenarioLine(
	findScenario func(id string) *messages.Scenario,
	findExample func(id string) (*messages.Examples, *messages.TableRow),
	pickle *messages.Pickle,
	file *fileProc,
) *lineProc {
	scenario := findScenario(pickle.AstNodeIds[0])

	result := f.Storage.MustGetPickleResult(pickle.Id)

	scenarioLine := findScenarioLine(file, scenario.Id)
	scenarioLine.StepResult = &stepResult{
		StartedAt: &result.StartedAt,
	}

	//revive:disable:add-constant
	if len(pickle.AstNodeIds) == 1 {
		return scenarioLine
	}
	//revive:enable:add-constant

	_, row := findExample(pickle.AstNodeIds[1])
	scenarioLine.ExampleLocation = row.Location

	return scenarioLine
}

func (f *htmlFmt) setStepsLine(
	findStep func(id string) *messages.Step,
	scenarioPickle *messages.Pickle,
	file *fileProc,
	scenarioLine *lineProc,
) {
	stepResults := f.Storage.MustGetPickleStepResultsByPickleID(
		scenarioPickle.Id,
	)

	for _, result := range stepResults {
		pickleStep := f.Storage.MustGetPickleStep(result.PickleStepID)
		step := findStep(pickleStep.AstNodeIds[0])

		stepLine := findStepLine(file, step.Id)

		startedAt := scenarioLine.StepResult.StartedAt
		finishAt := &result.FinishedAt
		duration := finishAt.Sub(*startedAt).Milliseconds()

		outputs := map[string]string{}
		for _, att := range result.Attachments {
			if att.MimeType == "text/plain" {
				outputs[att.Name] = string(att.Data)
			}
		}

		stepLine.StepResult = &stepResult{
			StartedAt:  startedAt,
			FinishedAt: finishAt,
			Duration:   duration,
			Status:     int(result.Status),
			Err:        result.Err,
			Outputs:    outputs,
		}
	}
}

func findScenarioLine(file *fileProc, scenarioID string) *lineProc {
	if file == nil {
		return nil
	}

	for _, line := range file.Lines {
		if line.LineType == LineTypeScenario && line.ID == scenarioID {
			return line
		}
	}

	return nil
}

func findStepLine(file *fileProc, stepID string) *lineProc {
	if file == nil {
		return nil
	}

	for _, line := range file.Lines {
		if line.LineType == LineTypeStep && line.ID == stepID {
			return line
		}
	}

	return nil
}

func complementLine(file *fileProc) {
	inheritLines := []LineType{
		LineTypeComment,
		LineTypeDocString,
		LineTypeTableHeader,
		LineTypeTableRow,
	}

	var preLine *lineProc
	for _, line := range file.Lines {
		line.file = file

		if slices.Contains(inheritLines, line.LineType) &&
			preLine != nil &&
			preLine.StepResult != nil {
			result := *preLine.StepResult
			result.Outputs = nil
			line.StepResult = &result
		}

		preLine = line
	}
}

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

func collectFileInDocument(doc *messages.GherkinDocument) *fileProc {
	if doc == nil {
		return nil
	}

	var lines []*lineProc

	fLines := collectLinesInFeature(doc.Feature)
	lines = append(lines, fLines...)

	for _, comment := range doc.Comments {
		commentLines := collectLinesInComment(comment)
		lines = append(lines, commentLines...)
	}

	//revive:disable
	sort.Sort(sortLinesByLocation(lines))
	//revive:enable

	return &fileProc{
		URI:   doc.Uri,
		Lines: lines,
	}
}

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

type lineProc struct {
	file            *fileProc
	LineType        LineType
	ID              string
	Location        messages.Location
	Tags            []*messages.Tag
	Keyword         string
	Name            string
	Description     string
	Text            string
	StepResult      *stepResult
	ExampleLocation *messages.Location
}

type stepResult struct {
	StartedAt  *time.Time
	FinishedAt *time.Time
	Duration   int64
	Status     int
	Err        error
	Outputs    map[string]string
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

	if l.ExampleLocation != nil {
		text += fmt.Sprintf(" (Example L.%d)", l.ExampleLocation.Line)
	}

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

type sortLinesByLocation []*lineProc

func (s sortLinesByLocation) Len() int {
	return len(s)
}

func (s sortLinesByLocation) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortLinesByLocation) Less(i, j int) bool {
	return s[i].Location.Line < s[j].Location.Line
}

func collectLinesInComment(comment *messages.Comment) []*lineProc {
	var lines []*lineProc

	if comment == nil {
		return lines
	}

	lines = append(lines, &lineProc{
		LineType: LineTypeComment,
		Location: *comment.Location,
		Text:     comment.Text,
	})

	return lines
}

func collectLinesInTag(tag *messages.Tag) []*lineProc {
	var lines []*lineProc

	if tag == nil {
		return lines
	}

	lines = append(lines, &lineProc{
		LineType: LineTypeTag,
		ID:       tag.Id,
		Location: *tag.Location,
		Text:     tag.Name,
	})

	return lines
}

func collectLinesInTags(tags []*messages.Tag) []*lineProc {
	var lines []*lineProc

	for _, tag := range tags {
		tagLines := collectLinesInTag(tag)
		lines = append(lines, tagLines...)
	}

	return lines
}

func collectLinesInFeature(feature *messages.Feature) []*lineProc {
	var lines []*lineProc

	if feature == nil {
		return lines
	}

	lines = append(lines, &lineProc{
		LineType:    LineTypeFeature,
		Location:    *feature.Location,
		Keyword:     feature.Keyword,
		Name:        feature.Name,
		Description: feature.Description,
	})

	tagLines := collectLinesInTags(feature.Tags)
	lines = append(lines, tagLines...)

	for _, child := range feature.Children {
		ruleLines := collectLinesInRule(child.Rule)
		lines = append(lines, ruleLines...)

		bgLines := collectLinesInBackground(child.Background)
		lines = append(lines, bgLines...)

		scenarioLine := collectLinesInScenario(child.Scenario)
		lines = append(lines, scenarioLine...)
	}

	return lines
}

func collectLinesInBackground(bg *messages.Background) []*lineProc {
	var lines []*lineProc

	if bg == nil {
		return lines
	}

	lines = append(lines, &lineProc{
		LineType:    LineTypeBackground,
		ID:          bg.Id,
		Location:    *bg.Location,
		Keyword:     bg.Keyword,
		Name:        bg.Name,
		Description: bg.Description,
	})

	stepLines := collectLinesInSteps(bg.Steps)
	lines = append(lines, stepLines...)

	return lines
}

func collectLinesInRule(rule *messages.Rule) []*lineProc {
	var lines []*lineProc

	if rule == nil {
		return lines
	}

	lines = append(lines, &lineProc{
		LineType:    LineTypeRule,
		ID:          rule.Id,
		Location:    *rule.Location,
		Keyword:     rule.Keyword,
		Name:        rule.Name,
		Description: rule.Description,
	})

	tagLines := collectLinesInTags(rule.Tags)
	lines = append(lines, tagLines...)

	for _, child := range rule.Children {
		bgLines := collectLinesInBackground(child.Background)
		lines = append(lines, bgLines...)

		scenarioLine := collectLinesInScenario(child.Scenario)
		lines = append(lines, scenarioLine...)
	}

	return lines
}

func collectLinesInScenario(scenario *messages.Scenario) []*lineProc {
	var lines []*lineProc

	if scenario == nil {
		return lines
	}

	lines = append(lines, &lineProc{
		LineType:    LineTypeScenario,
		ID:          scenario.Id,
		Location:    *scenario.Location,
		Tags:        scenario.Tags,
		Keyword:     scenario.Keyword,
		Name:        scenario.Name,
		Description: scenario.Description,
	})

	tagLines := collectLinesInTags(scenario.Tags)
	lines = append(lines, tagLines...)

	stepLines := collectLinesInSteps(scenario.Steps)
	lines = append(lines, stepLines...)

	exampleLines := collectLinesInExamples(scenario.Examples)
	lines = append(lines, exampleLines...)

	return lines
}

func collectLinesInSteps(steps []*messages.Step) []*lineProc {
	var lines []*lineProc

	for _, step := range steps {
		stepLines := collectLinesInStep(step)
		lines = append(lines, stepLines...)
	}

	return lines
}

func collectLinesInStep(step *messages.Step) []*lineProc {
	var lines []*lineProc

	if step == nil {
		return lines
	}

	lines = append(lines, &lineProc{
		LineType: LineTypeStep,
		ID:       step.Id,
		Location: *step.Location,
		Keyword:  step.Keyword,
		Text:     step.Text,
	})

	docStringLines := collectLinesInDocString(step.DocString)
	lines = append(lines, docStringLines...)

	if step.DataTable != nil {
		tableLines := collectLinesInTableRows(
			step.DataTable.Rows,
			LineTypeTableRow,
		)
		lines = append(lines, tableLines...)
	}

	return lines
}

func collectLinesInDocString(doc *messages.DocString) []*lineProc {
	var lines []*lineProc

	if doc == nil {
		return lines
	}

	// before """
	before := *doc.Location
	lines = append(lines, &lineProc{
		LineType: LineTypeDocString,
		Location: before,
		Text:     "\"\"\"",
	})

	// data
	loc := before
	for i, line := range strings.Split(doc.Content, "\n") {
		loc = before
		loc.Line += int64(i + 1)
		lines = append(lines, &lineProc{
			LineType: LineTypeDocString,
			Location: loc,
			Text:     line,
		})
	}

	// after """
	after := loc
	after.Line++
	lines = append(lines, &lineProc{
		LineType: LineTypeDocString,
		Location: after,
		Text:     "\"\"\"",
	})

	// TODO DataTable

	return lines
}

func collectLinesInExamples(examples []*messages.Examples) []*lineProc {
	var lines []*lineProc

	for _, example := range examples {
		egLines := collectLinesInExample(example)
		lines = append(lines, egLines...)
	}

	return lines
}

func collectLinesInExample(examples *messages.Examples) []*lineProc {
	var lines []*lineProc

	if examples == nil {
		return lines
	}

	lines = append(lines, &lineProc{
		LineType: LineTypeExamples,
		ID:       examples.Id,
		Location: *examples.Location,
		Keyword:  examples.Keyword,
		Name:     examples.Name,
	})

	header := collectLinesInTableRow(examples.TableHeader, LineTypeTableHeader)
	lines = append(lines, header...)

	body := collectLinesInTableRows(examples.TableBody, LineTypeTableRow)
	lines = append(lines, body...)

	return lines
}

func collectLinesInTableRows(
	rows []*messages.TableRow,
	lineType LineType,
) []*lineProc {
	var lines []*lineProc

	for _, row := range rows {
		r := collectLinesInTableRow(row, lineType)
		lines = append(lines, r...)
	}

	return lines
}

func collectLinesInTableRow(
	row *messages.TableRow,
	lineType LineType,
) []*lineProc {
	var lines []*lineProc

	if row == nil {
		return lines
	}

	values := []string{}
	for _, c := range row.Cells {
		values = append(values, c.Value)
	}

	lines = append(lines, &lineProc{
		LineType: lineType,
		Location: *row.Location,
		Text:     fmt.Sprintf("| %s |", strings.Join(values, " | ")),
	})

	return lines
}

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
