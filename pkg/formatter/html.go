package formatter

import (
	_ "embed"
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
			scenarioLine := f.setScenarioLine(
				feature.FindScenario,
				scenarioPickle,
				file,
			)
			f.setStepsLine(
				feature.FindStep,
				scenarioPickle,
				file,
				scenarioLine,
			)
		}
	}

	err := marshalHTML(files, f.out)
	if err != nil {
		panic(err)
	}
}

func (f *htmlFmt) setScenarioLine(
	findScenario func(id string) *messages.Scenario,
	pickle *messages.Pickle,
	file *fileProc,
) *lineProc {
	scenario := findScenario(pickle.AstNodeIds[0])

	result := f.Storage.MustGetPickleResult(pickle.Id)

	scenarioLine := findScenarioLine(file, scenario.Id)
	scenarioLine.StepResult = &stepResult{
		StartedAt: &result.StartedAt,
	}
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

		stepLine.StepResult = &stepResult{
			StartedAt:  startedAt,
			FinishedAt: finishAt,
			Duration:   duration,
			Status:     int(result.Status),
			Err:        result.Err,
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

type fileProc struct {
	URI   string
	Lines []*lineProc
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
	LineTypeComment    LineType = "comment"
	LineTypeTag        LineType = "tag"
	LineTypeFeature    LineType = "feature"
	LineTypeRule       LineType = "rule"
	LineTypeBackground LineType = "background"
	LineTypeScenario   LineType = "scenario"
	LineTypeStep       LineType = "step"
	LineTypeDocString  LineType = "docstring"
	LineTypeExamples   LineType = "examples"
)

type lineProc struct {
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

type stepResult struct {
	StartedAt  *time.Time
	FinishedAt *time.Time
	Duration   int64
	Status     int
	Err        error
}

func (l *lineProc) HTMLClass() string {
	classess := []string{string(l.LineType)}

	if l.StepResult != nil {
		//revive:disable:add-constant
		switch l.StepResult.Status {
		case 0:
			classess = append(classess, "passed")
		case 1:
			classess = append(classess, "failed")
		case 2:
			classess = append(classess, "skipped")
		case 3:
			classess = append(classess, "undefined")
		case 4:
			classess = append(classess, "pending")
		case 5:
			classess = append(classess, "ambiguous")
		default:
			classess = append(classess, "unknown")
		}
		//revive:enable:add-constant
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
			text += "\n"
		}
	}

	text += l.Name
	text += l.Text

	return text
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

	// TODO: DataTable

	return lines
}

func collectLinesInDocString(doc *messages.DocString) []*lineProc {
	var lines []*lineProc

	if doc == nil {
		return lines
	}

	lines = append(lines, &lineProc{
		LineType: LineTypeDocString,
		Location: *doc.Location,
		Text:     doc.Content,
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

	// TODO: TableHeader / TableBody

	return lines
}

//go:embed html.tmpl
var htmlTmpl string

func marshalHTML(files []*fileProc, out io.Writer) error {
	t, err := template.New("gherkin").Parse(htmlTmpl)
	if err != nil {
		return err
	}

	err = t.Execute(out, files)
	if err != nil {
		return err
	}

	return nil
}
