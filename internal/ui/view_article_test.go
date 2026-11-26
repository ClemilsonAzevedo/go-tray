package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	fyneTest "fyne.io/fyne/v2/test"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/news"
)

func TestArticleRowsKeepFixedDimensions(t *testing.T) {
	app := fyneTest.NewApp()
	app.Settings().SetTheme(NewTheme())
	t.Cleanup(app.Quit)

	articles := []news.Article{
		{Title: "Short title"},
		{
			Title:   strings.Repeat("A long headline that must be clamped ", 12),
			Summary: strings.Repeat("A long summary that must fit in two lines ", 12),
		},
	}
	for _, article := range articles {
		row := articleRow(article, func(string) {})
		if got, want := row.MinSize(), fyne.NewSize(articleRowWidth, articleCardHeight); got != want {
			t.Errorf("row size = %v, want %v", got, want)
		}
		if article.Summary != "" {
			wrapper := row.(*fyne.Container)
			card := wrapper.Objects[0].(*tappable).content.(*fyne.Container)
			paddedContent := card.Objects[1]
			if got := paddedContent.MinSize().Height; got > articleCardHeight {
				t.Errorf("article content height = %v, exceeds fixed card height %v", got, articleCardHeight)
			}
		}
	}
}

func TestArticleTextClampsAtTwoLines(t *testing.T) {
	app := fyneTest.NewApp()
	app.Settings().SetTheme(NewTheme())
	t.Cleanup(app.Quit)

	style := fyne.TextStyle{Bold: true}
	text := strings.Repeat("A headline with enough words to fill the fixed article width ", 12)
	got := clampTextLines(text, articleTextWidth, 2, articleTitleTextSize, style)
	lines := strings.Split(got, "\n")
	if len(lines) > 2 {
		t.Fatalf("clamped title has %d lines, want at most 2: %q", len(lines), got)
	}
	if !strings.Contains(got, "…") {
		t.Fatalf("long title should end in an ellipsis, got %q", got)
	}
	for _, line := range lines {
		if width := measuredTextWidth(line, articleTitleTextSize, style); width > articleTextWidth {
			t.Errorf("line width = %v, exceeds %v: %q", width, articleTextWidth, line)
		}
	}
}

func TestArticleFieldsFitWithinTwoLinesAndCard(t *testing.T) {
	app := fyneTest.NewApp()
	app.Settings().SetTheme(NewTheme())
	t.Cleanup(app.Quit)

	row := articleRow(news.Article{
		Source:  strings.Repeat("Notícias internacionais ", 12),
		Title:   strings.Repeat("A headline with enough words to fill the fixed article width ", 12),
		Summary: strings.Repeat("A summary with enough words to fill the fixed article width ", 12),
	}, func(string) {})
	wrapper := row.(*fyne.Container)
	card := wrapper.Objects[0].(*tappable).content.(*fyne.Container)
	paddedContent := card.Objects[1].(*fyne.Container)
	body := paddedContent.Objects[0].(*fyne.Container)
	header := body.Objects[0].(*fyne.Container)
	fields := map[string]*fyne.Container{
		"source":      header.Objects[0].(*fyne.Container),
		"title":       body.Objects[1].(*fyne.Container),
		"description": body.Objects[2].(*fyne.Container),
	}
	maxWidths := map[string]float32{
		"source":      articleTextWidth - header.Objects[2].(*canvas.Text).MinSize().Width - 2*articleHeaderSpacing - articleMetadataDotSize,
		"title":       articleTextWidth,
		"description": articleTextWidth,
	}
	if got, want := fields["source"].MinSize().Width, maxWidths["source"]; got != want {
		t.Errorf("source flex width = %v, want remaining header width %v", got, want)
	}
	for name, field := range fields {
		if len(field.Objects) > 2 {
			t.Errorf("%s renders %d separate lines, want at most 2", name, len(field.Objects))
		}
		for _, row := range field.Objects {
			line := row.(*fyne.Container).Objects[0].(*canvas.Text)
			if strings.Contains(line.Text, "\n") {
				t.Errorf("%s line still contains a newline: %q", name, line.Text)
			}
			if width := measuredTextWidth(line.Text, line.TextSize, line.TextStyle); width > maxWidths[name] {
				t.Errorf("%s line width = %v, exceeds %v: %q", name, width, maxWidths[name], line.Text)
			}
		}
	}
	for name, field := range fields {
		if len(field.Objects) != 2 {
			t.Errorf("long %s renders %d lines, want 2", name, len(field.Objects))
		}
	}
	lastSourceLine := fields["source"].Objects[1].(*fyne.Container).Objects[0].(*canvas.Text).Text
	if !strings.Contains(lastSourceLine, "…") {
		t.Errorf("long source should be truncated with an ellipsis, got %q", lastSourceLine)
	}
	if got := paddedContent.MinSize().Height; got > articleCardHeight {
		t.Errorf("article content height = %v, exceeds fixed card height %v", got, articleCardHeight)
	}
}

func TestArticleFieldsAlignShortTextToBottomOfTwoLineSlots(t *testing.T) {
	app := fyneTest.NewApp()
	app.Settings().SetTheme(NewTheme())
	t.Cleanup(app.Quit)

	row := articleRow(news.Article{
		Source:  "Source",
		Title:   "Short title",
		Summary: "Short description",
	}, func(string) {})
	wrapper := row.(*fyne.Container)
	card := wrapper.Objects[0].(*tappable).content.(*fyne.Container)
	body := card.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container)
	header := body.Objects[0].(*fyne.Container)
	fields := map[string]*fyne.Container{
		"source":      header.Objects[0].(*fyne.Container),
		"title":       body.Objects[1].(*fyne.Container),
		"description": body.Objects[2].(*fyne.Container),
	}
	for name, field := range fields {
		if len(field.Objects) != 2 {
			t.Errorf("short %s renders %d lines, want 2 fixed line slots", name, len(field.Objects))
			continue
		}
		firstLine := field.Objects[0].(*fyne.Container).Objects[0].(*canvas.Text).Text
		secondLine := field.Objects[1].(*fyne.Container).Objects[0].(*canvas.Text).Text
		if firstLine != "" {
			t.Errorf("short %s first line = %q, want an empty reserved line", name, firstLine)
		}
		if secondLine == "" {
			t.Errorf("short %s second line is empty, want the text aligned to the bottom", name)
		}
	}
}

func TestPlainTextPreservesPortugueseAndEnglishCharacters(t *testing.T) {
	input := "Notícias de ação — café &amp; English news"
	if got, want := plainText(input), "Notícias de ação — café & English news"; got != want {
		t.Fatalf("plainText() = %q, want %q", got, want)
	}

	if got := plainText(string([]byte{'n', 'e', 'w', 's', ' ', 0xff})); got != "news �" {
		t.Fatalf("plainText() should replace invalid UTF-8, got %q", got)
	}
}

func TestArticleSourceUsesLargerAccentTypography(t *testing.T) {
	if articleSourceTextSize != 11.5 {
		t.Errorf("source text size = %v, want 11.5", articleSourceTextSize)
	}
	if articleTitleTextSize != 13 {
		t.Errorf("title text size = %v, want 13", articleTitleTextSize)
	}
	if articleSummaryTextSize != 11.5 {
		t.Errorf("summary text size = %v, want 11.5", articleSummaryTextSize)
	}
}

func TestToolbarHoverOnlyBrightensCircle(t *testing.T) {
	for _, danger := range []bool{false, true} {
		normal := toolbarButtonColor(danger, false)
		hovered := toolbarButtonColor(danger, true)
		if normal.R != hovered.R || normal.G != hovered.G || normal.B != hovered.B {
			t.Errorf("danger=%v: hover changed circle hue from %v to %v", danger, normal, hovered)
		}
		if hovered.A <= normal.A {
			t.Errorf("danger=%v: hover alpha = %d, want greater than %d", danger, hovered.A, normal.A)
		}
	}
}
