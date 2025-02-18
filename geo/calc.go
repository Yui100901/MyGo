package geo

import (
	"math"
)

//
// @Author yfy2001
// @Date 2024/8/23 19 08
//

//部分地理相关计算函数

// EarthRadius 地球平均半径
const EarthRadius = 6371_393

// EarthCircumference 地球周长
const EarthCircumference = 2 * math.Pi * EarthRadius

// ExecOffset 传入原点p
// 传入经度和维度偏移
// 返回新的点的坐标
func ExecOffset(p *Coordinate, cf *CoordinateOffset) *Coordinate {
	newLongitude := p.Longitude + cf.LongitudeOffset
	newLatitude := p.Latitude + cf.LatitudeOffset
	return NewCoordinate(newLongitude, newLatitude)
}

func CalculateUnitDistances(c *Coordinate) *UnitDistances {

	latitudeRadians := c.LatitudeRadians

	//单位经度距离
	unitLongitudeDistance := EarthCircumference * math.Cos(latitudeRadians) / 360
	//单位纬度距离
	unitLatitudeDistance := EarthCircumference / 360

	return &UnitDistances{
		UnitLongitudeDistance: unitLongitudeDistance,
		UnitLatitudeDistance:  unitLatitudeDistance,
	}
}

// CalcCoordinateOffset 计算坐标偏移
// 传入原点
// 传入一个方位角上的距离偏移
// 返回坐标偏移量
func CalcCoordinateOffset(c *Coordinate, of *AzimuthOffset) *CoordinateOffset {

	//方向角弧度
	azimuthRadians := of.AzimuthRadians

	unitDistances := CalculateUnitDistances(c)

	//单位距离的经度变化量
	unitDeltaLongitude := 1 * math.Sin(azimuthRadians) / unitDistances.UnitLongitudeDistance
	//单位距离经度变化量
	unitDeltaLatitude := 1 * math.Cos(azimuthRadians) / unitDistances.UnitLatitudeDistance

	//经度偏移
	deltaLongitude := unitDeltaLongitude * of.Distance
	//纬度偏移
	deltaLatitude := unitDeltaLatitude * of.Distance

	return NewCoordinateOffset(deltaLongitude, deltaLatitude)

}

// DegreeToRadians 角度转弧度
func DegreeToRadians(degree float64) float64 {
	return degree * math.Pi / 180
}

// RadiansToDegree 弧度转角度
func RadiansToDegree(radians float64) float64 {
	return radians * 180 / math.Pi
}

// Haversine 公式，计算两点间距离
func Haversine(p1, p2 *Coordinate) float64 {

	//经度变化量
	dLongitudeRadians := p2.LongitudeRadians - p1.LongitudeRadians
	//纬度变化量
	dLatitudeRadians := p2.LatitudeRadians - p2.LatitudeRadians

	a := math.Sin(dLatitudeRadians/2)*math.Sin(dLatitudeRadians/2) +
		math.Cos(p1.Latitude*math.Pi/180.0)*
			math.Cos(p2.Latitude*math.Pi/180.0)*
			math.Sin(dLongitudeRadians/2)*math.Sin(dLongitudeRadians/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadius * c
}
