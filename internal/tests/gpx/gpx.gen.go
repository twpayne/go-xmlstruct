package gpx

import (
	"encoding/xml"
	"time"
)

type Gpx struct {
	XMLName        xml.Name `xml:"gpx"`
	Creator        string   `xml:"creator,attr"`
	SchemaLocation string   `xml:"schemaLocation,attr"`
	Topografix     *string  `xml:"topografix,attr"`
	Version        float64  `xml:"version,attr"`
	Xmlns          string   `xml:"xmlns,attr"`
	Xsi            string   `xml:"xsi,attr"`
	Active_point   *struct {
		Lat float64 `xml:"lat,attr"`
		Lon float64 `xml:"lon,attr"`
	} `xml:"active_point"`
	Author *string `xml:"author"`
	Bounds *struct {
		Maxlat float64 `xml:"maxlat,attr"`
		Maxlon float64 `xml:"maxlon,attr"`
		Minlat float64 `xml:"minlat,attr"`
		Minlon float64 `xml:"minlon,attr"`
	} `xml:"bounds"`
	Desc     *string `xml:"desc"`
	Email    *string `xml:"email"`
	Keywords *string `xml:"keywords"`
	Metadata struct {
		Author struct {
			Email struct {
				Domain string `xml:"domain,attr"`
				Id     string `xml:"id,attr"`
			} `xml:"email"`
			Link struct {
				Href string `xml:"href,attr"`
				Text string `xml:"text"`
			} `xml:"link"`
			Name string `xml:"name"`
		} `xml:"author"`
		Copyright struct {
			Author  string `xml:"author,attr"`
			License string `xml:"license"`
			Year    int    `xml:"year"`
		} `xml:"copyright"`
		Desc     string `xml:"desc"`
		Keywords string `xml:"keywords"`
		Link     struct {
			Href string `xml:"href,attr"`
			Text string `xml:"text"`
			Type string `xml:"type"`
		} `xml:"link"`
		Name string    `xml:"name"`
		Time time.Time `xml:"time"`
	} `xml:"metadata"`
	Name *string `xml:"name"`
	Rte  []struct {
		Desc string `xml:"desc"`
		Link *struct {
			Href string `xml:"href,attr"`
			Text string `xml:"text"`
		} `xml:"link"`
		Name   string `xml:"name"`
		Number string `xml:"number"`
		Rtept  []struct {
			Lat  float64  `xml:"lat,attr"`
			Lon  float64  `xml:"lon,attr"`
			Cmt  *string  `xml:"cmt"`
			Desc string   `xml:"desc"`
			Ele  *float64 `xml:"ele"`
			Link *struct {
				Href string `xml:"href,attr"`
				Text string `xml:"text"`
			} `xml:"link"`
			Name string     `xml:"name"`
			Sym  string     `xml:"sym"`
			Time *time.Time `xml:"time"`
			Type string     `xml:"type"`
		} `xml:"rtept"`
	} `xml:"rte"`
	Time *time.Time `xml:"time"`
	Trk  []*struct {
		Color *string `xml:"color"`
		Desc  string  `xml:"desc"`
		Link  *struct {
			Href string `xml:"href,attr"`
			Text string `xml:"text"`
		} `xml:"link"`
		Name   string `xml:"name"`
		Number string `xml:"number"`
		Trkseg struct {
			Trkpt []struct {
				Lat  float64    `xml:"lat,attr"`
				Lon  float64    `xml:"lon,attr"`
				Ele  *float64   `xml:"ele"`
				Sym  string     `xml:"sym"`
				Time *time.Time `xml:"time"`
			} `xml:"trkpt"`
		} `xml:"trkseg"`
	} `xml:"trk"`
	Url     *string `xml:"url"`
	Urlname *string `xml:"urlname"`
	Wpt     []struct {
		Lat  float64  `xml:"lat,attr"`
		Lon  float64  `xml:"lon,attr"`
		Cmt  *string  `xml:"cmt"`
		Desc string   `xml:"desc"`
		Ele  *float64 `xml:"ele"`
		Link *struct {
			Href string `xml:"href,attr"`
			Text string `xml:"text"`
		} `xml:"link"`
		Name *string    `xml:"name"`
		Sym  string     `xml:"sym"`
		Time *time.Time `xml:"time"`
		Type *string    `xml:"type"`
	} `xml:"wpt"`
}
