package models

type URLFrontierStatus int16

const (
	URL_FRONTIER_STATUS_NEW     URLFrontierStatus = iota
	URL_FRONTIER_STATUS_CRAWLED URLFrontierStatus = iota
	URL_FRONTIER_STATUS_ERROR   URLFrontierStatus = iota
)
