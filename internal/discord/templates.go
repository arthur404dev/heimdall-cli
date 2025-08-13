package discord

// Discord CSS template for most clients (Vesktop, Vencord, OpenAsar, Armcord)
const DiscordCSSTemplate = `/* Heimdall theme for Discord */
/* Generated automatically */

:root {
    /* Primary colors */
    --primary: #{{colour4}};
    --primary-container: #{{colour12}};
    --on-primary: #{{background}};
    --on-primary-container: #{{foreground}};
    
    /* Secondary colors */
    --secondary: #{{colour5}};
    --secondary-container: #{{colour13}};
    --on-secondary: #{{background}};
    
    /* Background colors */
    --background: #{{background}};
    --surface: #{{colour0}};
    --surface-variant: #{{colour8}};
    
    /* Text colors */
    --on-background: #{{foreground}};
    --on-surface: #{{foreground}};
    --on-surface-variant: #{{colour7}};
    
    /* Border colors */
    --outline: #{{colour8}};
    --outline-variant: #{{colour7}};
    
    /* Status colors */
    --error: #{{colour1}};
    --on-error: #{{background}};
    --success: #{{colour2}};
    --warning: #{{colour3}};
}

/* Discord specific mappings */
.theme-dark {
    /* Main backgrounds */
    --background-primary: var(--background);
    --background-secondary: var(--surface);
    --background-secondary-alt: var(--surface);
    --background-tertiary: var(--surface-variant);
    --background-accent: var(--primary);
    --background-floating: var(--surface);
    --background-mobile-primary: var(--background);
    --background-mobile-secondary: var(--surface);
    --background-modifier-hover: rgba({{colour4}}, 0.1);
    --background-modifier-active: rgba({{colour4}}, 0.2);
    --background-modifier-selected: rgba({{colour4}}, 0.3);
    --background-modifier-accent: rgba({{colour4}}, 0.4);
    
    /* Text colors */
    --text-normal: var(--on-background);
    --text-muted: var(--on-surface-variant);
    --text-faint: var(--outline);
    --text-positive: var(--success);
    --text-warning: var(--warning);
    --text-danger: var(--error);
    --text-brand: var(--primary);
    --text-link: var(--primary);
    
    /* Interactive elements */
    --interactive-normal: var(--on-surface);
    --interactive-hover: var(--primary);
    --interactive-active: var(--primary-container);
    --interactive-muted: var(--outline);
    
    /* Brand colors */
    --brand-experiment: var(--primary);
    --brand-experiment-hover: var(--primary-container);
    --brand-experiment-560: var(--primary);
    
    /* Channel colors */
    --channels-default: var(--on-surface-variant);
    --channel-icon: var(--on-surface-variant);
    --channel-text-area-placeholder: var(--outline);
    
    /* Header colors */
    --header-primary: var(--on-background);
    --header-secondary: var(--on-surface-variant);
    
    /* Scrollbar */
    --scrollbar-auto-thumb: var(--outline);
    --scrollbar-auto-track: transparent;
    --scrollbar-thin-thumb: var(--outline);
    --scrollbar-thin-track: transparent;
    
    /* Activity colors */
    --activity-card-background: var(--surface);
    
    /* Deprecated but still used */
    --deprecated-card-bg: var(--surface);
    --deprecated-card-editable-bg: var(--surface-variant);
    --deprecated-store-bg: var(--background);
    --deprecated-quickswitcher-input-background: var(--surface);
    --deprecated-quickswitcher-input-placeholder: var(--outline);
    --deprecated-text-input-bg: var(--surface);
    --deprecated-text-input-border: var(--outline);
    --deprecated-text-input-border-hover: var(--primary);
    --deprecated-text-input-border-disabled: var(--outline-variant);
    --deprecated-text-input-prefix: var(--on-surface-variant);
}

/* Light theme support */
.theme-light {
    /* Main backgrounds */
    --background-primary: #{{foreground}};
    --background-secondary: #{{colour15}};
    --background-secondary-alt: #{{colour15}};
    --background-tertiary: #{{colour7}};
    --background-accent: var(--primary);
    --background-floating: #{{colour15}};
    --background-mobile-primary: #{{foreground}};
    --background-mobile-secondary: #{{colour15}};
    --background-modifier-hover: rgba({{colour4}}, 0.1);
    --background-modifier-active: rgba({{colour4}}, 0.2);
    --background-modifier-selected: rgba({{colour4}}, 0.3);
    --background-modifier-accent: rgba({{colour4}}, 0.4);
    
    /* Text colors */
    --text-normal: #{{background}};
    --text-muted: #{{colour8}};
    --text-faint: #{{colour7}};
    --text-positive: var(--success);
    --text-warning: var(--warning);
    --text-danger: var(--error);
    --text-brand: var(--primary);
    --text-link: var(--primary);
    
    /* Interactive elements */
    --interactive-normal: #{{colour8}};
    --interactive-hover: var(--primary);
    --interactive-active: var(--primary-container);
    --interactive-muted: #{{colour7}};
    
    /* Brand colors */
    --brand-experiment: var(--primary);
    --brand-experiment-hover: var(--primary-container);
    --brand-experiment-560: var(--primary);
    
    /* Channel colors */
    --channels-default: #{{colour8}};
    --channel-icon: #{{colour8}};
    --channel-text-area-placeholder: #{{colour7}};
    
    /* Header colors */
    --header-primary: #{{background}};
    --header-secondary: #{{colour8}};
    
    /* Scrollbar */
    --scrollbar-auto-thumb: #{{colour7}};
    --scrollbar-auto-track: transparent;
    --scrollbar-thin-thumb: #{{colour7}};
    --scrollbar-thin-track: transparent;
    
    /* Activity colors */
    --activity-card-background: #{{colour15}};
    
    /* Deprecated but still used */
    --deprecated-card-bg: #{{colour15}};
    --deprecated-card-editable-bg: #{{colour7}};
    --deprecated-store-bg: #{{foreground}};
    --deprecated-quickswitcher-input-background: #{{colour15}};
    --deprecated-quickswitcher-input-placeholder: #{{colour7}};
    --deprecated-text-input-bg: #{{colour15}};
    --deprecated-text-input-border: #{{colour7}};
    --deprecated-text-input-border-hover: var(--primary);
    --deprecated-text-input-border-disabled: #{{colour8}};
    --deprecated-text-input-prefix: #{{colour8}};
}
`

// BetterDiscord theme template with META header
const BetterDiscordTemplate = `/**
 * @name Heimdall
 * @author heimdall-cli
 * @version 1.0.0
 * @description Heimdall color scheme for BetterDiscord
 * @source https://github.com/arthur404dev/heimdall-cli
 */

/* Heimdall theme for BetterDiscord */
/* Generated automatically */

:root {
    /* Primary colors */
    --primary: #{{colour4}};
    --primary-container: #{{colour12}};
    --on-primary: #{{background}};
    --on-primary-container: #{{foreground}};
    
    /* Secondary colors */
    --secondary: #{{colour5}};
    --secondary-container: #{{colour13}};
    --on-secondary: #{{background}};
    
    /* Background colors */
    --background: #{{background}};
    --surface: #{{colour0}};
    --surface-variant: #{{colour8}};
    
    /* Text colors */
    --on-background: #{{foreground}};
    --on-surface: #{{foreground}};
    --on-surface-variant: #{{colour7}};
    
    /* Border colors */
    --outline: #{{colour8}};
    --outline-variant: #{{colour7}};
    
    /* Status colors */
    --error: #{{colour1}};
    --on-error: #{{background}};
    --success: #{{colour2}};
    --warning: #{{colour3}};
}

/* Discord specific mappings for BetterDiscord */
.theme-dark {
    /* Main backgrounds */
    --background-primary: var(--background);
    --background-secondary: var(--surface);
    --background-secondary-alt: var(--surface);
    --background-tertiary: var(--surface-variant);
    --background-accent: var(--primary);
    --background-floating: var(--surface);
    --background-mobile-primary: var(--background);
    --background-mobile-secondary: var(--surface);
    --background-modifier-hover: rgba({{colour4}}, 0.1);
    --background-modifier-active: rgba({{colour4}}, 0.2);
    --background-modifier-selected: rgba({{colour4}}, 0.3);
    --background-modifier-accent: rgba({{colour4}}, 0.4);
    
    /* Text colors */
    --text-normal: var(--on-background);
    --text-muted: var(--on-surface-variant);
    --text-faint: var(--outline);
    --text-positive: var(--success);
    --text-warning: var(--warning);
    --text-danger: var(--error);
    --text-brand: var(--primary);
    --text-link: var(--primary);
    
    /* Interactive elements */
    --interactive-normal: var(--on-surface);
    --interactive-hover: var(--primary);
    --interactive-active: var(--primary-container);
    --interactive-muted: var(--outline);
    
    /* Brand colors */
    --brand-experiment: var(--primary);
    --brand-experiment-hover: var(--primary-container);
    --brand-experiment-560: var(--primary);
    
    /* Channel colors */
    --channels-default: var(--on-surface-variant);
    --channel-icon: var(--on-surface-variant);
    --channel-text-area-placeholder: var(--outline);
    
    /* Header colors */
    --header-primary: var(--on-background);
    --header-secondary: var(--on-surface-variant);
    
    /* Scrollbar */
    --scrollbar-auto-thumb: var(--outline);
    --scrollbar-auto-track: transparent;
    --scrollbar-thin-thumb: var(--outline);
    --scrollbar-thin-track: transparent;
    
    /* Activity colors */
    --activity-card-background: var(--surface);
    
    /* BetterDiscord specific */
    --bd-blue: var(--primary);
    --bd-blue-hover: var(--primary-container);
    --bd-blue-active: var(--primary-container);
}

/* BetterDiscord plugin/theme list styling */
.bd-addon-list .bd-addon-card {
    background-color: var(--surface);
    border-color: var(--outline);
}

.bd-addon-list .bd-addon-card:hover {
    background-color: var(--surface-variant);
}

.bd-addon-list .bd-addon-header {
    color: var(--on-surface);
}

.bd-addon-list .bd-addon-description {
    color: var(--on-surface-variant);
}

/* BetterDiscord settings styling */
.bd-settings-sidebar .bd-settings-item {
    color: var(--on-surface-variant);
}

.bd-settings-sidebar .bd-settings-item:hover {
    background-color: var(--surface-variant);
    color: var(--on-surface);
}

.bd-settings-sidebar .bd-settings-item.selected {
    background-color: var(--primary);
    color: var(--on-primary);
}
`

// GetTemplate returns the appropriate template for a Discord client
func GetTemplate(templateType string) string {
	switch templateType {
	case "betterdiscord":
		return BetterDiscordTemplate
	case "css":
		return DiscordCSSTemplate
	default:
		return DiscordCSSTemplate
	}
}
