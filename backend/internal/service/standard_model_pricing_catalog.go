package service

// StandardModelPricingCatalogEntry is a model exposed by both the administrator
// and user pricing pages. The list intentionally stays small: it tracks the
// models commonly selected in ChatGPT and Claude Code rather than every model
// present in the upstream catalog.
type StandardModelPricingCatalogEntry struct {
	Model    string
	Label    string
	Platform string
}

// StandardModelPricingCatalogGroup keeps the supplier grouping and its stable
// display order in one place.
type StandardModelPricingCatalogGroup struct {
	Name     string
	Platform string
	Models   []StandardModelPricingCatalogEntry
}

var standardModelPricingCatalog = []StandardModelPricingCatalogGroup{
	{
		Name:     "OpenAI GPT 常用模型",
		Platform: PlatformOpenAI,
		Models: []StandardModelPricingCatalogEntry{
			{Model: "gpt-5.5", Label: "GPT-5.5", Platform: PlatformOpenAI},
			{Model: "gpt-5.6-sol", Label: "GPT-5.6 Sol", Platform: PlatformOpenAI},
			{Model: "gpt-5.6-terra", Label: "GPT-5.6 Terra", Platform: PlatformOpenAI},
			{Model: "gpt-5.6-luna", Label: "GPT-5.6 Luna", Platform: PlatformOpenAI},
			{Model: "gpt-6-astra", Label: "GPT-6 Astra", Platform: PlatformOpenAI},
			{Model: "gpt-6-sol", Label: "GPT-6 Sol", Platform: PlatformOpenAI},
			{Model: "gpt-6-luna", Label: "GPT-6 Luna", Platform: PlatformOpenAI},
			{Model: "gpt-6.1-sol", Label: "GPT-6.1 Sol", Platform: PlatformOpenAI},
		},
	},
	{
		Name:     "Anthropic Claude Code 常用模型",
		Platform: PlatformAnthropic,
		Models: []StandardModelPricingCatalogEntry{
			{Model: "claude-opus-5-5", Label: "Claude Opus 5.5", Platform: PlatformAnthropic},
			{Model: "claude-opus-5", Label: "Claude Opus 5", Platform: PlatformAnthropic},
			{Model: "claude-sonnet-5-5", Label: "Claude Sonnet 5.5", Platform: PlatformAnthropic},
			{Model: "claude-sonnet-5", Label: "Claude Sonnet 5", Platform: PlatformAnthropic},
			{Model: "claude-opus-4-8", Label: "Claude Opus 4.8", Platform: PlatformAnthropic},
			{Model: "claude-opus-4-7", Label: "Claude Opus 4.7", Platform: PlatformAnthropic},
			{Model: "claude-opus-4-6", Label: "Claude Opus 4.6", Platform: PlatformAnthropic},
			{Model: "claude-sonnet-4-6", Label: "Claude Sonnet 4.6", Platform: PlatformAnthropic},
			{Model: "claude-haiku-4-5", Label: "Claude Haiku 4.5", Platform: PlatformAnthropic},
		},
	},
}

// StandardModelPricingCatalog returns a defensive copy so callers cannot
// modify the server-wide canonical model selection.
func StandardModelPricingCatalog() []StandardModelPricingCatalogGroup {
	out := make([]StandardModelPricingCatalogGroup, len(standardModelPricingCatalog))
	for i := range standardModelPricingCatalog {
		out[i] = standardModelPricingCatalog[i]
		out[i].Models = append([]StandardModelPricingCatalogEntry(nil), standardModelPricingCatalog[i].Models...)
	}
	return out
}
