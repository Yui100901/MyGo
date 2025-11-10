package geo

//
// @Author yfy2001
// @Date 2024/8/23 22 41
//

// Coordinate 坐标
type Coordinate struct {
	Longitude        float64 `json:"longitude"`        //经度
	Latitude         float64 `json:"latitude"`         //纬度
	LongitudeRadians float64 `json:"longitudeRadians"` //弧度制经度
	LatitudeRadians  float64 `json:"latitudeRadians"`  //弧度制纬度
	Altitude         float64 `json:"altitude"`         //海拔
	Height           float64 `json:"height"`           //相对高度
}

func NewCoordinate(longitude, latitude float64) *Coordinate {
	return &Coordinate{
		Longitude:        longitude,
		Latitude:         latitude,
		LongitudeRadians: DegreeToRadians(longitude),
		LatitudeRadians:  DegreeToRadians(latitude),
	}
}

// UnitDistances 结构体，用于保存单位经度和纬度的距离
type UnitDistances struct {
	UnitLongitudeDistance float64 //单位经度距离
	UnitLatitudeDistance  float64 //单位纬度距离
}

// BearingOffset 方位偏移
type BearingOffset struct {
	Bearing        float64 `json:"bearing"`        //方位角
	BearingRadians float64 `json:"bearingRadians"` //弧度制方位角
	Distance       float64 `json:"distance"`       //偏移距离
}

func NewBearingOffset(azimuth, distance float64) *BearingOffset {
	return &BearingOffset{
		Bearing:        azimuth,
		BearingRadians: DegreeToRadians(azimuth),
		Distance:       distance,
	}
}

// CoordinateOffset 坐标偏移
type CoordinateOffset struct {
	LongitudeOffset float64 `json:"longitudeOffset"` //经度偏移
	LatitudeOffset  float64 `json:"latitudeOffset"`  //纬度偏移
}

func NewCoordinateOffset(longitudeOffset, latitudeOffset float64) *CoordinateOffset {
	return &CoordinateOffset{
		LongitudeOffset: longitudeOffset,
		LatitudeOffset:  latitudeOffset,
	}
}
