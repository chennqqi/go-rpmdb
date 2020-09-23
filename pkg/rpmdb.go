package rpmdb

import (
	"github.com/chennqqi/go-rpmdb/pkg/bdb"
	"golang.org/x/xerrors"
)

type RpmDB struct {
	db *bdb.BerkeleyDB
}

func Open(path string) (*RpmDB, error) {
	db, err := bdb.Open(path)
	if err != nil {
		return nil, err
	}

	return &RpmDB{
		db: db,
	}, nil

}

func (d *RpmDB) ListPackages() ([]*PackageInfo, error) {
	var pkgList []*PackageInfo

	for entry := range d.db.Read() {
		if entry.Err != nil {
			return nil, entry.Err
		}

		indexEntries, err := headerImport(entry.Value)
		if err != nil {
			return nil, xerrors.Errorf("error during importing header: %w", err)
		}
		pkg, err := getNEVRA(indexEntries)
		if err != nil {
			return nil, xerrors.Errorf("invalid package info: %w", err)
		}
		pkgList = append(pkgList, pkg)
	}

	return pkgList, nil
}

/*
  -a, --all                        查询/验证所有软件包
  -f, --file                       查询/验证文件属于的软件包
  -g, --group                      查询/验证组中的软件包
  -p, --package                    查询/验证一个软件包

*/
func (d *RpmDB) ListPackagesWithTags(ids ...TAG_ID) ([]*PackageInfoEx, error) {
	var pkgList []*PackageInfoEx

	tagMask := make(map[TAG_ID]bool)
	for i := 0; i < len(ids); i++ {
		tagMask[ids[i]] = true
	}

	for entry := range d.db.Read() {
		if entry.Err != nil {
			return nil, entry.Err
		}

		indexEntries, err := headerImport(entry.Value)
		if err != nil {
			return nil, xerrors.Errorf("error during importing header: %w", err)
		}
		pkg, err := getPackageWithTags(indexEntries, tagMask)
		if err != nil {
			return nil, xerrors.Errorf("invalid package info: %w", err)
		}
		pkgList = append(pkgList, pkg)
	}

	return pkgList, nil
}
