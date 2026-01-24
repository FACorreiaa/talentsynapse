package adminpages

func GetRoleBadgeClass(role string) string {
	baseClass := "px-2 inline-flex text-xs leading-5 font-semibold rounded-full "
	colorClass := "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200"

	switch role {
	case "admin":
		colorClass = "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200"
	case "moderator":
		colorClass = "bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200"
	case "expert":
		colorClass = "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
	}

	return baseClass + colorClass
}
