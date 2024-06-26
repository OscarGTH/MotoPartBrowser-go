package main

import (
	"Crawler/internal/data"
	"Crawler/internal/database"
	"Crawler/internal/helpers"
	"Crawler/internal/models"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

const BrandReMatcher = "^[\\w-]+"

// Instantiates a Colly collector and configures it.
func createCollector() (*colly.Collector, error) {
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("purkuosat.net", "www.purkuosat.net"),
		colly.AllowURLRevisit(),
	)

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*purkuosat.*",
		Parallelism: 10,
		Delay:       50 * time.Millisecond,
		RandomDelay: 50 * time.Millisecond,
	})
	if err != nil {
		log.Fatalf("Cannot set limit rule. Reason: %s\n", err)
		return nil, err
	}
	return c, nil
}

func configureDefaultHandlers(c *colly.Collector) {
	// Set Fake User Agent and log visited URLs
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "1 Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148")
		log.Println("visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Println("Response received", r.StatusCode)
	})
}

func main() {
	helpers.ReadConfig()
	c, _ := createCollector()
	configureDefaultHandlers(c)

	var vehicles []models.RawVehicle

	// Each font element is a disassembled vehicle link
	c.OnHTML("font", func(e *colly.HTMLElement) {
		if e.Attr("size") == "2" {
			// Cloning the collector so we can visit the part page
			partCollector := c.Clone()

			// Instantiate a new vehicle and parts list for it.
			vehicle := models.RawVehicle{}
			var parts []models.RawPart

			// Grabbing basic info from the link text
			var vehicleName = e.ChildText("a")
			if len(vehicleName) == 0 {
				return
			}

			vehicle.Name = vehicleName
			vehicle.Url = e.Request.AbsoluteURL(e.ChildAttr("a", "href"))

			partCollector.OnRequest(func(r *colly.Request) {
				log.Println("Part collector visiting page:", r.URL.String())
			})

			partCollector.OnHTML("table", func(tb *colly.HTMLElement) {
				part := models.RawPart{}
				if tb.Attr("width") == "75%" {
					part.Name = tb.ChildText("tr:nth-of-type(1) > td:nth-of-type(3)")
					part.ImgThumbUrl = e.Request.AbsoluteURL(tb.ChildAttr("tr:nth-of-type(1) > td:nth-of-type(1) > a > img", "src"))
					part.ImgUrl = e.Request.AbsoluteURL(tb.ChildAttr("tr:nth-of-type(1) > td:nth-of-type(1) > a", "href"))
					part.PartIdentifier = tb.ChildText("tr:nth-of-type(2) > td:nth-of-type(2)")
					part.Description = tb.ChildText("tr:nth-of-type(3) > td:nth-of-type(2)")
					part.Price = tb.ChildText("tr:nth-of-type(4) > td:nth-of-type(2) > font > b:nth-of-type(1)")
					parts = append(parts, part)
				}
			})

			// Visiting the vehicle part page
			err := partCollector.Visit(vehicle.Url)
			if err != nil {
				log.Printf("Cannot visit the part page: %s. Reason: %s\n", vehicle.Url, err)
				return
			}

			// Add parts to the vehicle
			vehicle.RawParts = parts
			vehicles = append(vehicles, vehicle)
		} else {
			return
		}
	})

	// Retrieve the map of vehicle categories that should be crawled.
	categories := viper.GetStringMapString("crawl_categories")

	// Iterate over the vehicle categories.
	for category, listingPageUrl := range categories {
		var processedVehicles []models.Vehicle
		// Emptying vehicles slice when switching categories.
		vehicles = nil
		if viper.GetBool("loadFromJSON") {
			log.Println("Loading vehicles from JSON.")
			// Get the absolute path of the JSON file
			absJSONFilePath, err := filepath.Abs("./output/" + category + "_data.json")
			if err != nil {
				log.Fatal(err)
			}

			// Read vehicles from the JSON file
			processedVehicles, err = ReadVehiclesFromJSONFile(absJSONFilePath)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fName := "./output/" + category + "_data.json"
			file, err := os.Create(fName)
			if err != nil {
				log.Fatalf("Cannot create file %q: %s\n", fName, err)
				return
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Fatalf("Cannot close the file. Reason: %s\n", err)
					return
				}
			}(file)

			err = c.Visit(listingPageUrl)
			if err != nil {
				log.Fatalf("Cannot visit the page %s. Reason: %s\n", listingPageUrl, err)
				return
			}

			// Wait until all threads have finished.
			c.Wait()

			enc := json.NewEncoder(file)
			enc.SetIndent("", "  ")

			// Convert raw vehicles to vehicles
			processedVehicles := convertRawVehiclesToVehicles(vehicles, category)
			// Dump json to the standard output
			enc.Encode(processedVehicles)
			log.Printf("Successfully dumped json to the file %s.", file.Name())
		}
		log.Println("Connecting to database.")
		dbHandler := database.CreateDatabaseHandler()
		log.Println("Transfering vehicles to database.")
		transferVehiclesToDatabase(dbHandler, processedVehicles)
	}
}

// ReadVehiclesFromJSONFile reads vehicles from a JSON file and returns a slice of vehicles
func ReadVehiclesFromJSONFile(filePath string) ([]models.Vehicle, error) {
	// Read the JSON file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the JSON data into a slice of vehicles
	var vehicles []models.Vehicle
	err = json.NewDecoder(file).Decode(&vehicles)
	if err != nil {
		return nil, err
	}

	return vehicles, nil
}

// convertRawVehiclesToVehicles converts raw vehicle data to processed vehicle data
func convertRawVehiclesToVehicles(rawVehicles []models.RawVehicle, category string) []models.Vehicle {
	log.Printf("Converting %d raw vehicles to vehicles", len(rawVehicles))
	var vehicles []models.Vehicle
	for _, rawVehicle := range rawVehicles {
		vehicles = append(vehicles, processRawVehicle(rawVehicle, category))
	}
	return vehicles
}

// processVehicleData processes the raw vehicle data and returns a vehicle struct
func processRawVehicle(rawVehicle models.RawVehicle, category string) models.Vehicle {
	// Instantiate a new vehicle and parts list for it.
	var vehicle models.Vehicle
	var parts []models.Part

	vehicle.Name = standardizeSpaces(rawVehicle.Name)
	vehicle.Url = rawVehicle.Url
	vehicle.Year = extractYear(vehicle.Name)
	vehicle.Brand = extractBrand(vehicle.Name, BrandReMatcher)
	vehicle.Model = extractModel(vehicle.Name)
	vehicle.VehicleType = category
	vehicle.Identifier = generateHash(vehicle.Url, vehicle.Name)

	for _, part := range rawVehicle.RawParts {
		// Parsing price from string to float64
		price, err := parsePrice(part.Price)
		if err != nil {
			log.Printf("Cannot parse price string of part %s (Url: %s) (Value: %q). Reason: %s\n", part.PartIdentifier, vehicle.Url, part.Price, err)
		}
		// Making a new part.
		newPart := models.Part{
			Name:           standardizeSpaces(part.Name),
			Description:    standardizeSpaces(part.Description),
			PartIdentifier: generateHash(part.PartIdentifier, vehicle.Name),
			Price:          price,
			ImgUrl:         part.ImgUrl,
			ImgThumbUrl:    part.ImgThumbUrl,
		}
		parts = append(parts, newPart)
	}

	vehicle.Parts = parts
	return vehicle
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func parsePrice(price string) (float64, error) {
	// regular expression to match a sequence of digits optionally followed by a decimal point and more digits
	re := regexp.MustCompile(`(\d+\.\d+|\d+)`)
	matches := re.FindStringSubmatch(price)

	if len(matches) == 0 {
		return 0, fmt.Errorf("no price found in string")
	}

	priceFloat, err := strconv.ParseFloat(matches[0], 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse price: %w", err)
	}

	return priceFloat, nil
}

// extractYear extracts the year from a string
func extractYear(s string) int {
	// regular expression to match a sequence of digits optionally followed by a decimal point and more digits
	re := regexp.MustCompile(`\d{4}`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 0 {
		return 0
	}

	year, err := strconv.Atoi(matches[0])
	if err != nil {
		return 0
	}

	return year
}

// extractBrand extracts the brand from a string
func extractBrand(s string, regexMatcher string) string {
	// Trim the string
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(regexMatcher)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 0 {
		return ""
	}

	// Replace dashes with spaces in the matched brand
	match := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(matches[0], "-", " "), " ", ""))

	// Convert brands to a map for faster lookup
	brandsMap := make(map[string]string)
	for _, brand := range data.Brands {
		processedBrand := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(brand, "-", " "), " ", ""))
		brandsMap[processedBrand] = brand
	}

	// if the brand is in the map of brands, return it
	if brand, ok := brandsMap[match]; ok {
		return brand
	}

	// Recursively call the function with regex matcher that matches first two words.
	// Exit on the second call
	if regexMatcher == BrandReMatcher {
		return extractBrand(s, regexMatcher+"\\s[\\w-]+")
	}
	return ""
}

// extractModel extracts the model from a string
func extractModel(s string) string {
	// Trim the string
	s = strings.TrimSpace(s)
	// Model is between brand and year in the string, other can be discarded.
	// Identify brand first
	brand := extractBrand(s, BrandReMatcher)
	// If brand is not found, return empty string
	if len(brand) == 0 {
		return ""
	}
	// Identify year
	year := extractYear(s)

	// Identify the indexes of brand and year and take the string between them
	brandIndex := strings.Index(strings.ToLower(s), strings.ToLower(brand))
	// If year exists, then take the index of it
	var yearIndex int
	if year != 0 {
		yearIndex = strings.Index(s, strconv.Itoa(year))
		// Take the string between brand and year
		return strings.TrimSpace(s[brandIndex+len(brand) : yearIndex])
	} else {
		// If year does not exist, trim out the brand and take the rest of the string
		return strings.TrimSpace(s[brandIndex+len(brand):])
	}
}

// generateHash
// generates a hashed identifier from one or many string values.
func generateHash(hashableVals ...string) string {
	h := fnv.New32a()
	for _, a := range hashableVals {
		h.Write([]byte(a))
	}
	return fmt.Sprint(h.Sum32())
}

// transferVehiclesToDatabase writes the contents of the parsed vehicles and their parts there.
func transferVehiclesToDatabase(handler *database.PSQLHandler, vehicles []models.Vehicle) {
	// Verify the connection by pinging the database
	err := handler.DB.Ping()
	if err != nil {
		panic(err)
	}
	// Close connection after everything has been sent to database.
	defer handler.DB.Close()
	err = handler.InsertVehicles(vehicles)
	if err != nil {
		panic(err)
	}
	log.Printf("successfully inserted %d vehicles into database", len(vehicles))
	err = handler.InsertParts(vehicles)
	if err != nil {
		log.Fatalf("failed to insert parts to database %s", err)
	}
}
