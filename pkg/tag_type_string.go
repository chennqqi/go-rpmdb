// Code generated by "stringer -type=TAG_TYPE"; DO NOT EDIT.

package rpmdb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[RPM_NULL_TYPE-0]
	_ = x[RPM_CHAR_TYPE-1]
	_ = x[RPM_INT8_TYPE-2]
	_ = x[RPM_INT16_TYPE-3]
	_ = x[RPM_INT32_TYPE-4]
	_ = x[RPM_INT64_TYPE-5]
	_ = x[RPM_STRING_TYPE-6]
	_ = x[RPM_BIN_TYPE-7]
	_ = x[RPM_STRING_ARRAY_TYPE-8]
	_ = x[RPM_I18NSTRING_TYPE-9]
}

const _TAG_TYPE_name = "RPM_NULL_TYPERPM_CHAR_TYPERPM_INT8_TYPERPM_INT16_TYPERPM_INT32_TYPERPM_INT64_TYPERPM_STRING_TYPERPM_BIN_TYPERPM_STRING_ARRAY_TYPERPM_I18NSTRING_TYPE"

var _TAG_TYPE_index = [...]uint8{0, 13, 26, 39, 53, 67, 81, 96, 108, 129, 148}

func (i TAG_TYPE) String() string {
	if i >= TAG_TYPE(len(_TAG_TYPE_index)-1) {
		return "TAG_TYPE(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TAG_TYPE_name[_TAG_TYPE_index[i]:_TAG_TYPE_index[i+1]]
}
