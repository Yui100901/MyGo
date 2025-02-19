package geo

import "math"

//
// @Author yfy2001
// @Date 2025/2/19 09 36
//

const (
	pi  = math.Pi                // 圆周率（π）
	xPi = pi * 3000.0 / 180.0    // 特殊常量，用于百度坐标系的转换
	a   = 6378245.0              // 地球椭球体长半轴（单位：米），适用于GCJ-02坐标系
	ee  = 0.00669342162296594323 // 地球椭球体的偏心率平方，适用于GCJ-02坐标系
)

// 判断是否在中国境内
func outOfChina(lat, lon float64) bool {
	return !(lon > 73.66 && lon < 135.05 && lat > 3.86 && lat < 53.55)
}

// 经纬度转换函数 - 纬度
func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*pi) + 20.0*math.Sin(2.0*x*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*pi) + 40.0*math.Sin(y/3.0*pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*pi) + 320*math.Sin(y*pi/30.0)) * 2.0 / 3.0
	return ret
}

// 经纬度转换函数 - 经度
func transformLon(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*pi) + 20.0*math.Sin(2.0*x*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*pi) + 40.0*math.Sin(x/3.0*pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*pi) + 300.0*math.Sin(x/30.0*pi)) * 2.0 / 3.0
	return ret
}

// WGS84ToGCJ02 WGS84 转 GCJ02
func WGS84ToGCJ02(coord *Coordinate) *Coordinate {
	if outOfChina(coord.Latitude, coord.Longitude) {
		return coord
	}
	// 纬度转换
	dLat := transformLat(coord.Longitude-105.0, coord.Latitude-35.0)
	// 经度转换
	dLon := transformLon(coord.Longitude-105.0, coord.Latitude-35.0)
	// 计算弧度制纬度的正弦值
	radLat := coord.LatitudeRadians
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	// 计算弧度制纬度的平方根值
	sqrtMagic := math.Sqrt(magic)
	// 计算实际纬度差
	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * pi)
	// 计算实际经度差
	dLon = (dLon * 180.0) / (a / sqrtMagic * math.Cos(radLat) * pi)
	// 获取 GCJ-02 坐标
	mgLat := coord.Latitude + dLat
	mgLon := coord.Longitude + dLon
	return NewCoordinate(mgLon, mgLat)
}

// GCJ02ToWGS84 GCJ02 转 WGS84
func GCJ02ToWGS84(coord *Coordinate) *Coordinate {
	if outOfChina(coord.Latitude, coord.Longitude) {
		return coord
	}
	// 纬度转换
	dLat := transformLat(coord.Longitude-105.0, coord.Latitude-35.0)
	// 经度转换
	dLon := transformLon(coord.Longitude-105.0, coord.Latitude-35.0)
	// 计算弧度制纬度的正弦值
	radLat := coord.LatitudeRadians
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	// 计算弧度制纬度的平方根值
	sqrtMagic := math.Sqrt(magic)
	// 计算实际纬度差
	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * pi)
	// 计算实际经度差
	dLon = (dLon * 180.0) / (a / sqrtMagic * math.Cos(radLat) * pi)
	// 获取 WGS84 坐标
	mgLat := coord.Latitude + dLat
	mgLon := coord.Longitude + dLon
	return NewCoordinate(coord.Longitude*2-mgLon, coord.Latitude*2-mgLat)
}

// GCJ02ToBD09 GCJ02 转 BD09
func GCJ02ToBD09(coord *Coordinate) *Coordinate {
	x := coord.Longitude
	y := coord.Latitude
	// 计算 z 值
	z := math.Sqrt(x*x+y*y) + 0.00002*math.Sin(y*xPi)
	// 计算 θ 值
	theta := math.Atan2(y, x) + 0.000003*math.Cos(x*xPi)
	// 获取 BD09 坐标
	bdLon := z*math.Cos(theta) + 0.0065
	bdLat := z*math.Sin(theta) + 0.006
	return NewCoordinate(bdLon, bdLat)
}

// BD09ToGCJ02 BD09 转 GCJ02
func BD09ToGCJ02(coord *Coordinate) *Coordinate {
	x := coord.Longitude - 0.0065
	y := coord.Latitude - 0.006
	// 计算 z 值
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*xPi)
	// 计算 θ 值
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*xPi)
	// 获取 GCJ02 坐标
	ggLon := z * math.Cos(theta)
	ggLat := z * math.Sin(theta)
	return NewCoordinate(ggLon, ggLat)
}

// WGS84ToCGCS2000 WGS84 转 CGCS2000
func WGS84ToCGCS2000(coord *Coordinate) *Coordinate {
	// WGS84和CGCS2000坐标系非常接近，可以直接使用WGS84坐标
	return coord
}

// CGCS2000ToWGS84 CGCS2000 转 WGS84
func CGCS2000ToWGS84(coord *Coordinate) *Coordinate {
	// WGS84和CGCS2000坐标系非常接近，可以直接使用CGCS2000坐标
	return coord
}
