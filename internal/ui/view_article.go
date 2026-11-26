package ui

import (
	"fmt"
	"html"
	"image/color"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/news"
)

const (
	articleListHorizontalPadding = 6
	articleCardHorizontalPadding = 12
	articleCardVerticalPadding   = 5
	articleCardHeight            = 120
	articleSourceTextSize        = 11.5
	articleTitleTextSize         = 13
	articleSummaryTextSize       = 11.5
	articleFieldLineCount        = 2
	articleRowWidth              = popoverWidth - 2*articleListHorizontalPadding
	articleTextWidth             = articleRowWidth - 2*articleCardHorizontalPadding
	articleHeaderSpacing         = 6
	articleMetadataDotSize       = 2.5
)

func articleRow(article news.Article, open func(string)) fyne.CanvasObject {
	return articleRowForVariant(article, open, currentThemeVariant())
}

func articleRowForVariant(article news.Article, open func(string), variant fyne.ThemeVariant) fyne.CanvasObject {
	source := strings.ToUpper(plainText(article.Source))
	updated := metadataTextForVariant(ago(articleTime(article)), 10, false, articleMetadataSecondaryColorName, variant)
	sourceWidth := articleTextWidth - updated.MinSize().Width - 2*articleHeaderSpacing - articleMetadataDotSize
	sourceText := clampedText(source, sourceWidth, articleSourceTextSize, systemColorForVariant(articleAccentColorName, variant), fyne.TextStyle{Bold: true})
	dot := canvas.NewCircle(systemColorForVariant(articleMetadataTertiaryColorName, variant))
	dotIcon := container.NewCenter(container.NewGridWrap(fyne.NewSize(articleMetadataDotSize, articleMetadataDotSize), dot))
	header := container.New(layout.NewCustomPaddedHBoxLayout(articleHeaderSpacing), sourceText, dotIcon, updated)
	title := clampedText(plainText(article.Title), articleTextWidth, articleTitleTextSize, systemColorForVariant(theme.ColorNameForeground, variant), fyne.TextStyle{Bold: true})
	summary := clampedText(plainText(article.Summary), articleTextWidth, articleSummaryTextSize, systemColorForVariant(articleSummaryColorName, variant), fyne.TextStyle{})
	body := container.New(layout.NewCustomPaddedVBoxLayout(2), header, title, summary)

	background := canvas.NewRectangle(systemColorForVariant(theme.ColorNameBackground, variant))
	background.CornerRadius = 10
	card := container.NewStack(
		background,
		container.New(
			layout.NewCustomPaddedLayout(articleCardVerticalPadding, articleCardVerticalPadding, articleCardHorizontalPadding, articleCardHorizontalPadding),
			body,
		),
	)
	tappable := newTappable(card, background, func() { open(article.Link) })
	return container.New(layout.NewGridWrapLayout(fyne.NewSize(articleRowWidth, articleCardHeight)), tappable)
}

func ago(value time.Time) string {
	if value.IsZero() {
		return "data desconhecida"
	}

	elapsed := time.Since(value)
	switch {
	case elapsed < time.Minute:
		return "agora"
	case elapsed < time.Hour:
		return fmt.Sprintf("há %d min", int(elapsed.Minutes()))
	case elapsed < 24*time.Hour:
		return fmt.Sprintf("há %d h", int(elapsed.Hours()))
	default:
		return fmt.Sprintf("há %d d", int(elapsed.Hours()/24))
	}
}

func articleTime(article news.Article) time.Time {
	if !article.Date.IsZero() {
		return article.Date
	}
	for _, dateLayout := range []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822, time.RFC3339} {
		if parsed, err := time.Parse(dateLayout, article.Published); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

func plainText(value string) string {
	value = strings.ToValidUTF8(value, "�")
	value = htmlTagPattern.ReplaceAllString(value, " ")
	return strings.Join(strings.Fields(html.UnescapeString(value)), " ")
}

func clampTextLines(text string, maxWidth float32, maxLines int, size float32, style fyne.TextStyle) string {
	if maxWidth <= 0 || maxLines <= 0 {
		return ""
	}

	words := strings.Fields(text)
	lines := make([]string, 0, maxLines)
	current := ""
	for _, word := range words {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		if measuredTextWidth(candidate, size, style) <= maxWidth {
			current = candidate
			continue
		}

		if current == "" {
			return fitTextWithEllipsis(word, maxWidth, size, style)
		}

		lines = append(lines, current)
		if len(lines) == maxLines {
			lines[len(lines)-1] = ellipsizeText(lines[len(lines)-1], maxWidth, size, style)
			return strings.Join(lines, "\n")
		}
		if measuredTextWidth(word, size, style) > maxWidth {
			lines = append(lines, fitTextWithEllipsis(word, maxWidth, size, style))
			return strings.Join(lines, "\n")
		}
		current = word
	}

	if current != "" {
		lines = append(lines, current)
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[len(lines)-1] = ellipsizeText(lines[len(lines)-1], maxWidth, size, style)
	}
	return strings.Join(lines, "\n")
}

func clampedText(text string, maxWidth, size float32, textColor color.Color, style fyne.TextStyle) *fyne.Container {
	lines := strings.Split(clampTextLines(text, maxWidth, articleFieldLineCount, size, style), "\n")
	for len(lines) < articleFieldLineCount {
		lines = append([]string{""}, lines...)
	}

	measure := canvas.NewText("Mg", textColor)
	measure.TextSize = size
	measure.TextStyle = style
	lineHeight := measure.MinSize().Height

	rows := make([]fyne.CanvasObject, 0, articleFieldLineCount)
	for _, line := range lines {
		textLine := canvas.NewText(line, textColor)
		textLine.TextSize = size
		textLine.TextStyle = style
		rows = append(rows, container.NewGridWrap(fyne.NewSize(maxWidth, lineHeight), textLine))
	}
	return container.New(layout.NewCustomPaddedVBoxLayout(0), rows...)
}

func fitTextWithEllipsis(text string, maxWidth, size float32, style fyne.TextStyle) string {
	if measuredTextWidth(text, size, style) <= maxWidth {
		return text
	}
	return ellipsizeText(text, maxWidth, size, style)
}

func ellipsizeText(text string, maxWidth, size float32, style fyne.TextStyle) string {
	runes := []rune(text)
	for len(runes) > 0 {
		candidate := strings.TrimSpace(string(runes)) + "…"
		if measuredTextWidth(candidate, size, style) <= maxWidth {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	if measuredTextWidth("…", size, style) <= maxWidth {
		return "…"
	}
	return ""
}

func measuredTextWidth(text string, size float32, style fyne.TextStyle) float32 {
	measure := canvas.NewText(text, textMain())
	measure.TextSize = size
	measure.TextStyle = style
	return measure.MinSize().Width
}
