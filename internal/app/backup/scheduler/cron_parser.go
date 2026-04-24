package scheduler

import (
	"fmt"
	"strings"
)

// ConvertCronToSystemd translates a basic 5-column cron expression to systemd OnCalendar format.
func ConvertCronToSystemd(cronExpr string) (string, error) {
	fields := strings.Fields(strings.TrimSpace(cronExpr))
	if len(fields) != 5 {
		return "", fmt.Errorf("invalid cron expression: expected 5 fields, got %d", len(fields))
	}

	minute := fields[0]
	hour := fields[1]
	dom := fields[2]
	month := fields[3]
	dow := fields[4]

	// Systemd format: DayOfWeek Year-Month-Day Hour:Minute:Second

	// Convert Day of Week
	sysDow := ""
	if dow != "*" {
		sysDow = dow + " "
	}

	sysMonth := month
	sysDom := dom
	if strings.HasPrefix(dom, "*/") {
		step := strings.TrimPrefix(dom, "*/")
		sysDom = "1/" + step
	}

	sysHour := hour
	if hour != "*" && !strings.HasPrefix(hour, "*/") && len(sysHour) == 1 {
		sysHour = "0" + sysHour
	} else if strings.HasPrefix(hour, "*/") {
		step := strings.TrimPrefix(hour, "*/")
		sysHour = "0/" + step
	}

	sysMinute := minute
	if minute != "*" && !strings.HasPrefix(minute, "*/") && len(sysMinute) == 1 {
		sysMinute = "0" + sysMinute
	} else if strings.HasPrefix(minute, "*/") {
		step := strings.TrimPrefix(minute, "*/")
		sysMinute = "0/" + step
	}

	onCalendar := fmt.Sprintf("%s*-%s-%s %s:%s:00", sysDow, sysMonth, sysDom, sysHour, sysMinute)
	return onCalendar, nil
}
