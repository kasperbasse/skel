package cmd

// prettyHelpTemplate adds a bit more vertical spacing between standard Cobra help sections
const prettyHelpTemplate = `
{{with (or .Long .Short)}}{{.}}{{end}}

Usage:
  {{if .Runnable}}{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}{{.CommandPath}} [command]{{end}}

{{if gt (len .Aliases) 0}}Aliases:
  {{.NameAndAliases}}

{{end}}{{if .HasExample}}Examples:
{{.Example}}

{{end}}{{if .HasAvailableSubCommands}}Available Commands:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}

{{end}}{{if .HasAvailableLocalFlags}}Options:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}{{if .HasAvailableInheritedFlags}}Global Options:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}{{if .HasHelpSubCommands}}Additional help topics:
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}
{{end}}{{end}}

{{end}}{{if .HasAvailableSubCommands}}Use "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}`
