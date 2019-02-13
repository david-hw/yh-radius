package main

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
)

func parsePkg(pkg []byte) RadiusPackage {
	rp := RadiusPackage{}
	rp.Code = pkg[0]
	rp.Identifier = pkg[1]
	rp.Length = binary.BigEndian.Uint16(pkg[2:4])
	rp.Authenticator = getSixteenBytes(pkg[4:20])
	rp.AuthenticatorString = strings.ToUpper(hex.EncodeToString(rp.Authenticator[:]))
	// 解析radius属性
	attrs := make([]RadiusAttr, 0, 50)
	attrs = parseRadiusAttr(pkg[20:], attrs, &rp)
	rp.RadiusAttrs = attrs
	return rp
}

// 解析radius属性: type(1) + length(1) + value
func parseRadiusAttr(attrBytes []byte, attrs []RadiusAttr, rp *RadiusPackage)  []RadiusAttr {
	length := len(attrBytes)
	if length == 0 {
		return attrs
	}

	attrType := attrBytes[0]
	attrLength := attrBytes[1]
	attr := RadiusAttr{
		AttrType:   attrType,
		AttrLength: attrLength,
	}

	// 26号私有属性特殊处理
	if attrType == VENDOR_SPECIFIC_TYPE {
		attr.VendorId = binary.BigEndian.Uint32(attrBytes[ATTR_HEADER_LENGHT: ATTR_HEADER_LENGHT + 4])
		parseSpecRadiusAttr(attrBytes[VENDOR_HEADER_LENGTH:attrLength], &attr)
	} else {
		attr.AttrValue = attrBytes[ATTR_HEADER_LENGHT:attrLength]
		// 设置属性值的字符串形式值
		attribute, ok := ATTRITUBES[AttrKey{0, int(attrType)}]
		if ok {
			attr.AttrName = attribute.Name
		}
		attr.setStandardAttrStringVal()

		if attrType == 3 {
			rp.isChap = true
		}

		if attrType == 60 {
			rp.challenge = getSixteenBytes(attr.AttrValue)
		}
	}
	attrs = append(attrs, attr)
	attrs = parseRadiusAttr(attrBytes[attrLength:], attrs, rp)
	return attrs
}

// 解析厂商私有属性
func parseSpecRadiusAttr(specAttrBytes []byte, attr *RadiusAttr) {

	vendorType := specAttrBytes[0]
	vendorLength := specAttrBytes[1]

	vendorAttr := VendorAttr{
		VendorId: attr.VendorId,
		VendorType: vendorType,
		VendorLength: vendorLength,
		VendorValue: specAttrBytes[ATTR_HEADER_LENGHT: vendorLength],
	}
	// 设置属性值的字符串形式值
	vendorAttr.setVendorAttrStringValue()
	attribute, ok := ATTRITUBES[AttrKey{attr.VendorId, int(vendorType)}]
	if ok {
		vendorAttr.VendorTypeName = attribute.Name
	}

	attr.VendorAttrs = append(attr.VendorAttrs, vendorAttr)

	parseSpecRadiusAttr(specAttrBytes[vendorLength:], attr)
}

// 获取16字节Authenticator
func getSixteenBytes(source []byte) (bts [AUTHENTICATOR_LENGTH]byte) {
	for index, val := range source {
		if index >= AUTHENTICATOR_LENGTH {
			break
		}
		bts[index] = val
	}
	return
}