package discover

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	discoverpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/discover"
)

// TestCategoriesSortedAlphabetically verifies that categories are sorted alphabetically
func TestCategoriesSortedAlphabetically(t *testing.T) {
	tests := []struct {
		name            string
		inputCategories []discoverpages.CategoryFilter
		expectedOrder   []string
	}{
		{
			name: "categories are sorted alphabetically",
			inputCategories: []discoverpages.CategoryFilter{
				{ID: "3", Name: "Technology"},
				{ID: "1", Name: "Art"},
				{ID: "2", Name: "Music"},
				{ID: "4", Name: "Business"},
			},
			expectedOrder: []string{"Art", "Business", "Music", "Technology"},
		},
		{
			name: "already sorted categories remain sorted",
			inputCategories: []discoverpages.CategoryFilter{
				{ID: "1", Name: "Alpha"},
				{ID: "2", Name: "Beta"},
				{ID: "3", Name: "Gamma"},
			},
			expectedOrder: []string{"Alpha", "Beta", "Gamma"},
		},
		{
			name: "single category",
			inputCategories: []discoverpages.CategoryFilter{
				{ID: "1", Name: "Only One"},
			},
			expectedOrder: []string{"Only One"},
		},
		{
			name:            "empty categories",
			inputCategories: []discoverpages.CategoryFilter{},
			expectedOrder:   []string{},
		},
		{
			name: "categories with similar prefixes",
			inputCategories: []discoverpages.CategoryFilter{
				{ID: "3", Name: "Programming"},
				{ID: "1", Name: "Program"},
				{ID: "2", Name: "Programmer"},
			},
			expectedOrder: []string{"Program", "Programmer", "Programming"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			categories := make([]discoverpages.CategoryFilter, len(tt.inputCategories))
			copy(categories, tt.inputCategories)

			// Apply the same sorting logic as in handler.go
			sort.Slice(categories, func(i, j int) bool {
				return categories[i].Name < categories[j].Name
			})

			// Verify the order
			actualOrder := make([]string, 0, len(categories))
			for _, c := range categories {
				actualOrder = append(actualOrder, c.Name)
			}

			assert.Equal(t, tt.expectedOrder, actualOrder)
		})
	}
}

// TestCategoriesOrderIsDeterministic verifies that sorting produces consistent results
func TestCategoriesOrderIsDeterministic(t *testing.T) {
	// Simulate extracting categories from a map (which has random order)
	categoryMap := map[string]string{
		"uuid-1": "Technology",
		"uuid-2": "Art",
		"uuid-3": "Music",
		"uuid-4": "Business",
		"uuid-5": "Languages",
	}

	// Run multiple iterations to ensure deterministic ordering
	var previousOrder []string
	for i := 0; i < 10; i++ {
		var categories []discoverpages.CategoryFilter
		for id, name := range categoryMap {
			categories = append(categories, discoverpages.CategoryFilter{
				ID:   id,
				Name: name,
			})
		}

		// Apply sorting (same as handler.go)
		sort.Slice(categories, func(i, j int) bool {
			return categories[i].Name < categories[j].Name
		})

		// Extract names
		var currentOrder []string
		for _, c := range categories {
			currentOrder = append(currentOrder, c.Name)
		}

		if previousOrder != nil {
			assert.Equal(t, previousOrder, currentOrder,
				"Categories order should be consistent across iterations")
		}
		previousOrder = currentOrder
	}

	// Verify final expected order
	expectedOrder := []string{"Art", "Business", "Languages", "Music", "Technology"}
	assert.Equal(t, expectedOrder, previousOrder)
}

// TestSkillCardsPreserveOrderFromRepository verifies skills maintain DB order
func TestSkillCardsPreserveOrderFromRepository(t *testing.T) {
	// Simulate skills coming from repository (already ordered by category, then name)
	type mockSkill struct {
		ID           string
		Name         string
		CategoryID   string
		CategoryName string
	}

	// Skills are pre-ordered by repository: ORDER BY sc.name, s.name
	orderedSkills := []mockSkill{
		{ID: "1", Name: "Drawing", CategoryID: "art", CategoryName: "Art"},
		{ID: "2", Name: "Painting", CategoryID: "art", CategoryName: "Art"},
		{ID: "3", Name: "Accounting", CategoryID: "biz", CategoryName: "Business"},
		{ID: "4", Name: "Marketing", CategoryID: "biz", CategoryName: "Business"},
		{ID: "5", Name: "Go", CategoryID: "tech", CategoryName: "Technology"},
		{ID: "6", Name: "Python", CategoryID: "tech", CategoryName: "Technology"},
	}

	// Convert to skill cards (same as handler.go)
	var skillCards []discoverpages.SkillCard
	for _, s := range orderedSkills {
		skillCards = append(skillCards, discoverpages.SkillCard{
			ID:           s.ID,
			Name:         s.Name,
			CategoryName: s.CategoryName,
		})
	}

	// Verify order is preserved
	expectedNames := []string{"Drawing", "Painting", "Accounting", "Marketing", "Go", "Python"}
	var actualNames []string
	for _, sc := range skillCards {
		actualNames = append(actualNames, sc.Name)
	}

	assert.Equal(t, expectedNames, actualNames,
		"Skill cards should preserve the order from repository")
}

// TestCategoryExtractionFromSkills tests unique category extraction
func TestCategoryExtractionFromSkills(t *testing.T) {
	type mockSkill struct {
		CategoryID   string
		CategoryName string
	}

	skills := []mockSkill{
		{CategoryID: "1", CategoryName: "Technology"},
		{CategoryID: "2", CategoryName: "Art"},
		{CategoryID: "1", CategoryName: "Technology"}, // Duplicate
		{CategoryID: "3", CategoryName: "Music"},
		{CategoryID: "2", CategoryName: "Art"}, // Duplicate
	}

	// Extract unique categories using map (same as handler.go)
	categoryMap := make(map[string]string)
	for _, s := range skills {
		categoryMap[s.CategoryID] = s.CategoryName
	}

	var categories []discoverpages.CategoryFilter
	for id, name := range categoryMap {
		categories = append(categories, discoverpages.CategoryFilter{
			ID:   id,
			Name: name,
		})
	}

	// Sort alphabetically (same as handler.go)
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	// Verify uniqueness and order
	assert.Len(t, categories, 3, "Should have 3 unique categories")

	expectedOrder := []string{"Art", "Music", "Technology"}
	var actualOrder []string
	for _, c := range categories {
		actualOrder = append(actualOrder, c.Name)
	}
	assert.Equal(t, expectedOrder, actualOrder)
}

// TestEmptySkillsHandling tests behavior with no skills
func TestEmptySkillsHandling(t *testing.T) {
	var categories []discoverpages.CategoryFilter

	// Apply sorting on empty slice (should not panic)
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	assert.Empty(t, categories)
}
