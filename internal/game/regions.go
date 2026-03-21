package game

import "github.com/rhaqim/worldgame/internal/models"

// AllRegions defines the available regions for game creation.
// No base_stats -- regions are purely geographical context.
var AllRegions = []models.Region{
	{
		ID: "usa", Name: "United States", Country: "United States", Continent: "North America",
		Description: "Global superpower facing polarized politics, debates over tech regulation, rising inequality, and ongoing cybersecurity threats from state and non-state actors.",
	},
	{
		ID: "china", Name: "China", Country: "China", Continent: "Asia",
		Description: "Rising superpower navigating trade tensions, rapid AI advancement, demographic shifts, South China Sea disputes, and ambitious Belt and Road infrastructure expansion.",
	},
	{
		ID: "germany", Name: "Germany", Country: "Germany", Continent: "Europe",
		Description: "Europe's industrial engine grappling with energy transition, skilled labor shortages, regulatory complexity, and maintaining competitiveness amid global supply chain shifts.",
	},
	{
		ID: "japan", Name: "Japan", Country: "Japan", Continent: "Asia",
		Description: "Technological powerhouse contending with an aging population, economic stagnation pressures, regional security tensions, and pushing frontiers in robotics and green energy.",
	},
	{
		ID: "brazil", Name: "Brazil", Country: "Brazil", Continent: "South America",
		Description: "Resource-rich nation confronting Amazon deforestation debates, urban crime, political polarization, education inequality, and a booming agribusiness sector.",
	},
	{
		ID: "india", Name: "India", Country: "India", Continent: "Asia",
		Description: "World's most populous democracy managing rapid digital transformation, religious tensions, air quality crises, a massive youth bulge, and expanding startup ecosystem.",
	},
	{
		ID: "uk", Name: "United Kingdom", Country: "United Kingdom", Continent: "Europe",
		Description: "Post-Brexit power recalibrating trade relationships, managing NHS pressures, addressing regional inequality, and maintaining global intelligence and financial leadership.",
	},
	{
		ID: "france", Name: "France", Country: "France", Continent: "Europe",
		Description: "Nuclear power with strong social safety net facing pension reform protests, immigration debates, suburban unrest, and ambitions for European defense autonomy.",
	},
	{
		ID: "south_korea", Name: "South Korea", Country: "South Korea", Continent: "Asia",
		Description: "Tech-driven economy dealing with North Korean tensions, demographic decline, intense academic pressure, semiconductor competition, and cultural soft power expansion.",
	},
	{
		ID: "australia", Name: "Australia", Country: "Australia", Continent: "Oceania",
		Description: "Stable democracy facing climate-driven wildfires and droughts, mining-vs-environment tensions, Pacific security commitments, and indigenous reconciliation challenges.",
	},
	{
		ID: "russia", Name: "Russia", Country: "Russia", Continent: "Europe",
		Description: "Military power under international sanctions, managing economic isolation, Arctic resource ambitions, cyber-warfare capabilities, and internal governance pressures.",
	},
	{
		ID: "nigeria", Name: "Nigeria", Country: "Nigeria", Continent: "Africa",
		Description: "Africa's largest economy battling Boko Haram insurgency, oil-dependency, massive youth unemployment, cybercrime waves, and a vibrant but underfunded tech startup scene.",
	},
	{
		ID: "south_africa", Name: "South Africa", Country: "South Africa", Continent: "Africa",
		Description: "Most industrialized African nation tackling load-shedding energy crises, high crime rates, land reform debates, corruption fallout, and university access inequality.",
	},
	{
		ID: "mexico", Name: "Mexico", Country: "Mexico", Continent: "North America",
		Description: "Major manufacturing and nearshoring hub confronting cartel violence, judicial reform controversies, migration pressures, water scarcity, and growing middle class demands.",
	},
	{
		ID: "egypt", Name: "Egypt", Country: "Egypt", Continent: "Africa",
		Description: "Strategic gateway between Africa and the Middle East dealing with Suez Canal security, youth unemployment, human rights scrutiny, water disputes over the Nile, and economic reform pressures.",
	},
	{
		ID: "indonesia", Name: "Indonesia", Country: "Indonesia", Continent: "Asia",
		Description: "Southeast Asia's largest economy managing deforestation, digital economy growth, island connectivity challenges, religious pluralism, and new capital city construction.",
	},
	{
		ID: "saudi_arabia", Name: "Saudi Arabia", Country: "Saudi Arabia", Continent: "Asia",
		Description: "Oil-rich kingdom pursuing Vision 2030 diversification, social modernization, regional rivalries, labor market nationalization, and mega-project construction boom.",
	},
	{
		ID: "canada", Name: "Canada", Country: "Canada", Continent: "North America",
		Description: "Multicultural democracy navigating indigenous reconciliation, housing affordability crisis, Arctic sovereignty claims, immigration integration, and natural resource governance.",
	},
}

// GetRegionByID looks up a region by its ID.
func GetRegionByID(id string) *models.Region {
	for i := range AllRegions {
		if AllRegions[i].ID == id {
			return &AllRegions[i]
		}
	}
	return nil
}
