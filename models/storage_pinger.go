package models

import "github.com/coopernurse/gorp"

type StoragePinger struct {
	dbmapMaster *gorp.DbMap
	dbmapSlave  *gorp.DbMap
}

func NewStoragePinger(dbmapMaster *gorp.DbMap, dbmapSlave *gorp.DbMap) *StoragePinger {
	storagePinger := StoragePinger{dbmapMaster: dbmapMaster, dbmapSlave: dbmapSlave}
	return &storagePinger
}

func (storagePinger *StoragePinger) PingMaster() error {

	_, err := storagePinger.dbmapMaster.SelectInt("select 1")
	return err
}

func (storagePinger *StoragePinger) PingSlave() error {

	_, err := storagePinger.dbmapSlave.SelectInt("select 1")
	return err
}
