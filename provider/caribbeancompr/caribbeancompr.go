package caribbeancompr

import (
	"fmt"
	"regexp"

	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/caribbeancom"
	"github.com/javtube/javtube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*CaribbeancomPremium)(nil)

const (
	Name     = "CaribbeancomPR"
	Priority = 1000 - 1
)

const (
	baseURL  = "https://www.caribbeancompr.com/"
	movieURL = "https://www.caribbeancompr.com/moviepages/%s/index.html"
)

type CaribbeancomPremium struct {
	*caribbeancom.Caribbeancom
}

func New() *CaribbeancomPremium {
	return &CaribbeancomPremium{
		// Simply use Caribbeancom provider to scrape contents.
		Caribbeancom: &caribbeancom.Caribbeancom{
			Scraper:      scraper.NewDefaultScraper(Name, Priority, scraper.WithDetectCharset()),
			DefaultMaker: "カリビアンコムプレミアム",
		},
	}
}

func (carib *CaribbeancomPremium) NormalizeID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func (carib *CaribbeancomPremium) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return carib.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
