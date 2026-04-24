package scheduler

const serviceTemplate = `[Unit]
Description=sfDBTools Backup Job - {{ .JobName }}
After=network.target

[Service]
Type=oneshot
User=root
# We use flock to ensure serial execution (no parallel backup jobs running at the same time)
ExecStart=/usr/bin/flock /var/lock/sfdbtools-global-backup.lock {{ .ExecPath }} backup {{ .Mode }} --quiet --ticket="{{ .Ticket }}" --profile="{{ .Profile }}"{{ if .IncludeFile }} --db-file="{{ .IncludeFile }}"{{ end }}{{ if .OutputMode }} --mode="{{ .OutputMode }}"{{ end }} --backup-dir="{{ .OutputDir }}"
`

const timerTemplate = `[Unit]
Description=Timer for sfDBTools Backup Job - {{ .JobName }}

[Timer]
OnCalendar={{ .OnCalendar }}
Persistent=true

[Install]
WantedBy=timers.target
`
