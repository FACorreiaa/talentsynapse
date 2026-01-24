package calendar

// Helper functions for calendar templates

func prevMonth(month int) int {
	if month == 1 {
		return 12
	}
	return month - 1
}

func nextMonth(month int) int {
	if month == 12 {
		return 1
	}
	return month + 1
}

func prevYear(year int, month int) int {
	if month == 1 {
		return year - 1
	}
	return year
}

func nextYear(year int, month int) int {
	if month == 12 {
		return year + 1
	}
	return year
}

func sessionStatusClass(status string) string {
	switch status {
	case "pending":
		return "bg-yellow-100 text-yellow-800 border-yellow-300"
	case "confirmed", "scheduled":
		return "bg-blue-100 text-blue-800 border-blue-300"
	case "in_progress":
		return "bg-green-100 text-green-800 border-green-300"
	case "completed":
		return "bg-gray-100 text-gray-800 border-gray-300"
	case "cancelled", "no_show":
		return "bg-red-100 text-red-800 border-red-300"
	default:
		return "bg-gray-100 text-gray-800"
	}
}

func sessionStatusBadge(status string) string {
	switch status {
	case "pending":
		return "badge-warning"
	case "confirmed", "scheduled":
		return "badge-info"
	case "in_progress":
		return "badge-success"
	case "completed":
		return "badge-neutral"
	case "cancelled", "no_show":
		return "badge-error"
	default:
		return "badge-ghost"
	}
}
