package game

import "github.com/rhaqim/worldgame/internal/models"

var AllRegions = []models.Region{
	{
		ID: "usa", Name: "United States", Country: "United States", Continent: "North America",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 80, models.Politics: 60, models.Security: 75, models.Education: 70, models.RandD: 85,
		},
		Description: "Global superpower with a strong economy, advanced military, and world-leading R&D institutions.",
	},
	{
		ID: "china", Name: "China", Country: "China", Continent: "Asia",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 75, models.Politics: 70, models.Security: 70, models.Education: 60, models.RandD: 80,
		},
		Description: "Rising superpower with rapid economic growth, centralized governance, and heavy R&D investment.",
	},
	{
		ID: "germany", Name: "Germany", Country: "Germany", Continent: "Europe",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 78, models.Politics: 72, models.Security: 60, models.Education: 80, models.RandD: 82,
		},
		Description: "Industrial powerhouse of Europe with world-class engineering and strong education systems.",
	},
	{
		ID: "japan", Name: "Japan", Country: "Japan", Continent: "Asia",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 72, models.Politics: 65, models.Security: 55, models.Education: 78, models.RandD: 88,
		},
		Description: "Technological leader with a disciplined economy, high education standards, and cutting-edge R&D.",
	},
	{
		ID: "brazil", Name: "Brazil", Country: "Brazil", Continent: "South America",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 55, models.Politics: 45, models.Security: 40, models.Education: 45, models.RandD: 40,
		},
		Description: "Resource-rich nation with a growing economy but challenges in governance and security.",
	},
	{
		ID: "india", Name: "India", Country: "India", Continent: "Asia",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 60, models.Politics: 55, models.Security: 50, models.Education: 55, models.RandD: 65,
		},
		Description: "Rapidly growing economy with a large tech workforce and democratic governance.",
	},
	{
		ID: "uk", Name: "United Kingdom", Country: "United Kingdom", Continent: "Europe",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 72, models.Politics: 68, models.Security: 70, models.Education: 75, models.RandD: 78,
		},
		Description: "Historic global power with strong financial sector, intelligence capabilities, and elite universities.",
	},
	{
		ID: "france", Name: "France", Country: "France", Continent: "Europe",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 70, models.Politics: 62, models.Security: 68, models.Education: 72, models.RandD: 75,
		},
		Description: "Major European power with nuclear capability, strong culture of diplomacy, and advanced research.",
	},
	{
		ID: "south_korea", Name: "South Korea", Country: "South Korea", Continent: "Asia",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 70, models.Politics: 60, models.Security: 60, models.Education: 82, models.RandD: 85,
		},
		Description: "Tech-savvy economy with world-leading electronics industry and intense education culture.",
	},
	{
		ID: "australia", Name: "Australia", Country: "Australia", Continent: "Oceania",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 68, models.Politics: 70, models.Security: 55, models.Education: 72, models.RandD: 65,
		},
		Description: "Stable democracy with rich natural resources and a high quality of life.",
	},
	{
		ID: "russia", Name: "Russia", Country: "Russia", Continent: "Europe",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 50, models.Politics: 55, models.Security: 80, models.Education: 65, models.RandD: 60,
		},
		Description: "Military superpower with vast resources, strong security apparatus, and space heritage.",
	},
	{
		ID: "nigeria", Name: "Nigeria", Country: "Nigeria", Continent: "Africa",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 45, models.Politics: 40, models.Security: 35, models.Education: 40, models.RandD: 30,
		},
		Description: "Africa's largest economy with a young population and growing tech scene.",
	},
	{
		ID: "south_africa", Name: "South Africa", Country: "South Africa", Continent: "Africa",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 48, models.Politics: 50, models.Security: 40, models.Education: 50, models.RandD: 42,
		},
		Description: "Most industrialized African nation with strong mining sector and democratic institutions.",
	},
	{
		ID: "mexico", Name: "Mexico", Country: "Mexico", Continent: "North America",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 52, models.Politics: 45, models.Security: 38, models.Education: 48, models.RandD: 40,
		},
		Description: "Major manufacturing hub with rich culture and growing middle class.",
	},
	{
		ID: "egypt", Name: "Egypt", Country: "Egypt", Continent: "Africa",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 42, models.Politics: 45, models.Security: 50, models.Education: 45, models.RandD: 35,
		},
		Description: "Strategic gateway between Africa and Middle East with ancient heritage and regional influence.",
	},
	{
		ID: "indonesia", Name: "Indonesia", Country: "Indonesia", Continent: "Asia",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 50, models.Politics: 48, models.Security: 42, models.Education: 45, models.RandD: 35,
		},
		Description: "Southeast Asia's largest economy with vast archipelago and growing digital economy.",
	},
	{
		ID: "saudi_arabia", Name: "Saudi Arabia", Country: "Saudi Arabia", Continent: "Asia",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 65, models.Politics: 55, models.Security: 65, models.Education: 50, models.RandD: 45,
		},
		Description: "Oil-rich kingdom investing heavily in economic diversification and modernization.",
	},
	{
		ID: "canada", Name: "Canada", Country: "Canada", Continent: "North America",
		BaseStats: map[models.SectorType]float64{
			models.Economics: 70, models.Politics: 75, models.Security: 55, models.Education: 78, models.RandD: 72,
		},
		Description: "Stable democracy with strong social systems, natural resources, and multicultural society.",
	},
}

func GetRegionByID(id string) *models.Region {
	for i := range AllRegions {
		if AllRegions[i].ID == id {
			return &AllRegions[i]
		}
	}
	return nil
}
