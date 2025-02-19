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

// CalculateCoordinateOffset 计算坐标偏移
// 传入原点
// 传入一个方位角上的距离偏移
// 返回坐标偏移量
func CalculateCoordinateOffset(c *Coordinate, of *BearingOffset) *CoordinateOffset {

	//方向角弧度
	bearingRadians := of.BearingRadians

	unitDistances := CalculateUnitDistances(c)

	//单位距离的经度变化量
	unitDeltaLongitude := 1 * math.Sin(bearingRadians) / unitDistances.UnitLongitudeDistance
	//单位距离经度变化量
	unitDeltaLatitude := 1 * math.Cos(bearingRadians) / unitDistances.UnitLatitudeDistance

	//经度偏移
	deltaLongitude := unitDeltaLongitude * of.Distance
	//纬度偏移
	deltaLatitude := unitDeltaLatitude * of.Distance

	return NewCoordinateOffset(deltaLongitude, deltaLatitude)

}

// CalculateBearing Calculate initial CalculateBearing between two points
// 传入两个点坐标
// 计算方位角
func CalculateBearing(p1, p2 *Coordinate) float64 {
	// 计算两点经度差（单位：弧度）
	deltaLambda := p2.LongitudeRadians - p1.LongitudeRadians

	// 计算 y 和 x
	y := math.Sin(deltaLambda) * math.Cos(p2.LatitudeRadians)
	x := math.Cos(p1.LatitudeRadians)*math.Sin(p2.LatitudeRadians) -
		math.Sin(p1.LatitudeRadians)*math.Cos(p2.LatitudeRadians)*math.Cos(deltaLambda)

	// 使用 atan2 计算初始方位角（单位：弧度）
	theta := math.Atan2(y, x)

	// 将弧度转换为度数，并归一化到 0-360 度
	bearing := math.Mod(theta*180/math.Pi+360, 360)

	return bearing
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
