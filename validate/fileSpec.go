package validate

import (
	"github.com/hhrutter/pdfcpu/types"

	"github.com/pkg/errors"
)

// See 7.11.4

func validateEmbeddedFileStreamMacParameterDict(xRefTable *types.XRefTable, dict *types.PDFDict) (err error) {

	logInfoValidate.Println("*** validateEmbeddedFileStreamMacParameterDict begin ***")

	dictName := "embeddedFileStreamMacParameterDict"

	// Subtype, optional integer
	// The embedded file's file type integer encoded according to Mac OS conventions.
	_, err = validateIntegerEntry(xRefTable, dict, dictName, "Subtype", OPTIONAL, types.V10, nil)
	if err != nil {
		return
	}

	// Creator, optional integer
	// The embedded file's creator signature integer encoded according to Mac OS conventions.
	_, err = validateIntegerEntry(xRefTable, dict, dictName, "Creator", OPTIONAL, types.V10, nil)
	if err != nil {
		return
	}

	// ResFork, optional stream dict
	// The binary contents of the embedded file's resource fork.
	_, err = validateStreamDictEntry(xRefTable, dict, dictName, "ResFork", OPTIONAL, types.V10, nil)
	if err != nil {
		return
	}

	logInfoValidate.Println("*** validateEmbeddedFileStreamMacParameterDict end ***")

	return
}

func validateEmbeddedFileStreamParameterDict(xRefTable *types.XRefTable, obj interface{}) (err error) {

	logInfoValidate.Println("*** validateEmbeddedFileStreamParameterDict begin ***")

	dict, err := xRefTable.DereferenceDict(obj)
	if err != nil {
		return
	}

	if obj == nil {
		logInfoValidate.Println("validateEmbeddedFileStreamParameterDict end")
		return
	}

	dictName := "embeddedFileStreamParmDict"

	// Size, optional integer
	_, err = validateIntegerEntry(xRefTable, dict, dictName, "Size", OPTIONAL, types.V10, nil)
	if err != nil {
		return
	}

	// CreationDate, optional date
	_, err = validateDateEntry(xRefTable, dict, dictName, "CreationDate", OPTIONAL, types.V10)
	if err != nil {
		return
	}

	// ModDate, optional date
	_, err = validateDateEntry(xRefTable, dict, dictName, "ModDate", OPTIONAL, types.V10)
	if err != nil {
		return
	}

	// Mac, optional dict
	macDict, err := validateDictEntry(xRefTable, dict, dictName, "Mac", OPTIONAL, types.V10, nil)
	if err != nil {
		return
	}
	if macDict != nil {
		err = validateEmbeddedFileStreamMacParameterDict(xRefTable, macDict)
		if err != nil {
			return
		}
	}

	// CheckSum, optional string
	_, err = validateStringEntry(xRefTable, dict, dictName, "CheckSum", OPTIONAL, types.V10, nil)
	if err != nil {
		return
	}

	logInfoValidate.Println("*** validateEmbeddedFileStreamParameterDict end ***")

	return
}

func validateEmbeddedFileStreamDict(xRefTable *types.XRefTable, sd *types.PDFStreamDict) (err error) {

	logInfoValidate.Println("*** validateEmbeddedFileStreamDict begin ***")

	dictName := "embeddedFileStreamDict"

	// Type, optional, name
	_, err = validateNameEntry(xRefTable, &sd.PDFDict, dictName, "Type", OPTIONAL, types.V10, func(s string) bool { return s == "EmbeddedFile" })
	if err != nil {
		return
	}

	// Subtype, optional, name
	_, err = validateNameEntry(xRefTable, &sd.PDFDict, dictName, "Subtype", OPTIONAL, types.V10, nil)
	if err != nil {
		return
	}

	// Params, optional, dict
	// parameter dict containing additional file-specific information.
	if obj, found := sd.PDFDict.Find("Params"); found && obj != nil {
		err = validateEmbeddedFileStreamParameterDict(xRefTable, obj)
		if err != nil {
			return
		}
	}

	logInfoValidate.Println("*** validateEmbeddedFileStreamDict end ***")

	return
}

func validateFileSpecDictEntriesEFAndRFKeys(k string) bool {
	return k == "F" || k == "UF" || k == "DOS" || k == "Mac" || k == "Unix"
}

func validateFileSpecDictEntryEFDict(xRefTable *types.XRefTable, dict *types.PDFDict) (err error) {

	for k, obj := range (*dict).Dict {

		if !validateFileSpecDictEntriesEFAndRFKeys(k) {
			return errors.Errorf("validateFileSpecEntriesEFAndRF: invalid key: %s", k)
		}

		// value must be embedded file stream dict
		// see 7.11.4
		var sd *types.PDFStreamDict
		sd, err = validateStreamDict(xRefTable, obj)
		if err != nil {
			return
		}

		err = validateEmbeddedFileStreamDict(xRefTable, sd)
		if err != nil {
			return
		}

		continue

	}

	return
}

func validateEFDictFilesArray(xRefTable *types.XRefTable, arr *types.PDFArray) (err error) {

	if len(*arr)%2 > 0 {
		return errors.New("validateEFDictFilesArray: rfDict array corrupt")
	}

	for k, v := range *arr {

		if v == nil {
			return errors.New("validateEFDictFilesArray: rfDict, array entry nil")
		}

		var obj interface{}
		obj, err = xRefTable.Dereference(v)
		if err != nil {
			return
		}

		if obj == nil {
			return errors.New("validateEFDictFilesArray: rfDict, array entry nil")
		}

		if k%2 > 0 {

			_, ok := obj.(types.PDFStringLiteral)
			if !ok {
				return errors.New("validateEFDictFilesArray: rfDict, array entry corrupt")
			}

		} else {

			// value must be embedded file stream dict
			// see 7.11.4
			sd, err := validateStreamDict(xRefTable, obj)
			if err != nil {
				return err
			}

			err = validateEmbeddedFileStreamDict(xRefTable, sd)
			if err != nil {
				return err
			}

		}
	}

	return
}

func validateFileSpecDictEntriesEFAndRF(xRefTable *types.XRefTable, efDict, rfDict *types.PDFDict) (err error) {

	// EF only or EF and RF

	logInfoValidate.Println("*** validateFileSpecDictEntriesEFAndRF begin ***")

	if efDict == nil {
		return errors.Errorf("validateFileSpecEntriesEFAndRF: missing required efDict.")
	}

	err = validateFileSpecDictEntryEFDict(xRefTable, efDict)
	if err != nil {
		return
	}

	if rfDict != nil {

		for k, val := range (*rfDict).Dict {

			if _, ok := efDict.Find(k); !ok {
				return errors.Errorf("validateFileSpecEntriesEFAndRF: rfDict entry=%s missing corresponding efDict entry\n", k)
			}

			// value must be related files array.
			// see 7.11.4.2
			var arr *types.PDFArray
			arr, err = xRefTable.DereferenceArray(val)
			if err != nil {
				return
			}

			if arr == nil {
				continue
			}

			err = validateEFDictFilesArray(xRefTable, arr)
			if err != nil {
				return
			}

		}

	}

	logInfoValidate.Println("*** validateFileSpecDictEntriesEFAndRF end ***")

	return
}

func validateFileSpecDictType(xRefTable *types.XRefTable, dict *types.PDFDict) (err error) {

	if dict.Type() == nil || (*dict.Type() != "Filespec" && (xRefTable.ValidationMode == types.ValidationRelaxed && *dict.Type() != "F")) {
		return errors.New("validateFileSpecDictType: missing type: FileSpec")
	}
	return
}

func requiredF(dosFound, macFound, unixFound bool) bool {
	return !dosFound && !macFound && !unixFound
}

func validateFileSpecDictEFAndEF(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string) (err error) {

	// RF, optional, dict of related files arrays, since V1.3
	rfDict, err := validateDictEntry(xRefTable, dict, dictName, "RF", OPTIONAL, types.V13, nil)
	if err != nil {
		return
	}

	// EF, required if RF present, dict of embedded file streams, since 1.3
	efDict, err := validateDictEntry(xRefTable, dict, dictName, "EF", rfDict != nil, types.V13, nil)
	if err != nil {
		return
	}

	// Type, required if EF present, name
	validate := func(s string) bool {
		return s == "Filespec" || (xRefTable.ValidationMode == types.ValidationRelaxed && s == "F")
	}
	_, err = validateNameEntry(xRefTable, dict, dictName, "Type", efDict != nil, types.V10, validate)
	if err != nil {
		return
	}

	// if EF present, Type "FileSpec" is required
	if efDict != nil {

		err = validateFileSpecDictType(xRefTable, dict)
		if err != nil {
			return
		}

		err = validateFileSpecDictEntriesEFAndRF(xRefTable, efDict, rfDict)
		if err != nil {
			return
		}

	}

	return
}

func validateFileSpecDict(xRefTable *types.XRefTable, dict *types.PDFDict) (err error) {

	logInfoValidate.Println("*** validateFileSpecDict begin ***")

	dictName := "fileSpecDict"

	// FS, optional, name
	fsName, err := validateNameEntry(xRefTable, dict, dictName, "FS", OPTIONAL, types.V10, nil)
	if err != nil {
		return
	}

	// DOS, byte string, optional, obsolescent.
	_, dosFound := dict.Find("DOS")

	// Mac, byte string, optional, obsolescent.
	_, macFound := dict.Find("Mac")

	// Unix, byte string, optional, obsolescent.
	_, unixFound := dict.Find("Unix")

	// F, file spec string
	validate := validateFileSpecString
	if fsName != nil && fsName.Value() == "URL" {
		validate = validateURLString
	}

	_, err = validateStringEntry(xRefTable, dict, dictName, "F", requiredF(dosFound, macFound, unixFound), types.V10, validate)
	if err != nil {
		return
	}

	// UF, optional, text string
	sinceVersion := types.V17
	if xRefTable.ValidationMode == types.ValidationRelaxed {
		sinceVersion = types.V14
	}
	_, err = validateStringEntry(xRefTable, dict, dictName, "UF", OPTIONAL, sinceVersion, validateFileSpecString)
	if err != nil {
		return
	}

	// ID, optional, array of strings
	_, err = validateStringArrayEntry(xRefTable, dict, dictName, "ID", OPTIONAL, types.V11, func(arr types.PDFArray) bool { return len(arr) == 2 })
	if err != nil {
		return
	}

	// V, optional, boolean, since V1.2
	_, err = validateBooleanEntry(xRefTable, dict, dictName, "V", OPTIONAL, types.V12, nil)
	if err != nil {
		return
	}

	err = validateFileSpecDictEFAndEF(xRefTable, dict, dictName)
	if err != nil {
		return
	}

	// Desc, optional, text string, since V1.6
	sinceVersion = types.V16
	if xRefTable.ValidationMode == types.ValidationRelaxed {
		sinceVersion = types.V10
	}
	_, err = validateStringEntry(xRefTable, dict, dictName, "Desc", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return
	}

	// CI, optional, collection item dict, since V1.7
	_, err = validateDictEntry(xRefTable, dict, dictName, "CI", OPTIONAL, types.V17, nil)
	if err != nil {
		return
	}

	logInfoValidate.Println("*** validateFileSpecDict end ***")

	return
}

func validateFileSpecification(xRefTable *types.XRefTable, obj interface{}) (o interface{}, err error) {

	// See 7.11.4

	logInfoValidate.Println("*** validateFileSpecification begin ***")

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return
	}

	switch obj := obj.(type) {

	case types.PDFStringLiteral:
		s := obj.Value()
		if !validateFileSpecString(s) {
			err = errors.Errorf("validateFileSpecification: invalid file spec string: %s", s)
			return
		}

	case types.PDFHexLiteral:
		s := obj.Value()
		if !validateFileSpecString(s) {
			err = errors.Errorf("validateFileSpecification: invalid file spec string: %s", s)
			return
		}

	case types.PDFDict:
		err = validateFileSpecDict(xRefTable, &obj)
		if err != nil {
			return
		}

	default:
		err = errors.Errorf("validateFileSpecification: invalid type")
		return
	}

	o = obj

	logInfoValidate.Println("*** validateFileSpecification end ***")

	return
}

func validateFileSpecEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string,
	required bool, sinceVersion types.PDFVersion) (o interface{}, err error) {

	logInfoValidate.Printf("*** validateFileSpecEntry begin: entry=%s ***\n", entryName)

	obj, found := dict.Find(entryName)
	if !found || obj == nil {
		if required {
			err = errors.Errorf("validateFileSpecEntry: dict=%s required entry=%s missing", dictName, entryName)
			return
		}
		logInfoValidate.Printf("validateFileSpecEntry end: entry %s is nil\n", entryName)
		return
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return
	}

	if obj == nil {
		if required {
			err = errors.Errorf("validateFileSpecEntry: dict=%s required entry=%s missing", dictName, entryName)
			return
		}
		logInfoValidate.Printf("validateFileSpecEntry end: entry %s is nil\n", entryName)
		return
	}

	o, err = validateFileSpecification(xRefTable, obj)
	if err != nil {
		return
	}

	logInfoValidate.Printf("*** validateFileSpecEntry end: entry=%s ***\n", entryName)

	return
}

func validateFileSpecificationOrFormObject(xRefTable *types.XRefTable, obj interface{}) (o interface{}, err error) {

	logInfoValidate.Println("*** validateFileSpecificationOrFormObject begin ***")

	switch obj := obj.(type) {

	case types.PDFStringLiteral:
		s := obj.Value()
		if !validateFileSpecString(s) {
			err = errors.Errorf("validateFileSpecificationOrFormObject: invalid file spec string: %s", s)
			return
		}

	case types.PDFHexLiteral:
		s := obj.Value()
		if !validateFileSpecString(s) {
			err = errors.Errorf("validateFileSpecificationOrFormObject: invalid file spec string: %s", s)
			return
		}

	case types.PDFDict:
		err = validateFileSpecDict(xRefTable, &obj)
		if err != nil {
			return
		}

	case types.PDFStreamDict:
		err = validateFormStreamDict(xRefTable, &obj)
		if err != nil {
			return
		}

	default:
		err = errors.Errorf("validateFileSpecificationOrFormObject: invalid type")
		return
	}

	o = obj

	logInfoValidate.Println("*** validateFileSpecificationOrFormObject end ***")

	return
}
