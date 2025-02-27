package provider

import (
	"errors"
	"fmt"
	"regexp"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/request"
)

func (w *Website) GetModules() ([]model.Module, error) {
	var modules []model.Module
	err := w.DB.Order("id asc").Find(&modules).Error
	if err != nil {
		return nil, err
	}
	return modules, nil
}

func (w *Website) GetModuleById(id uint) (*model.Module, error) {
	var module model.Module
	db := w.DB
	err := db.Where("`id` = ?", id).First(&module).Error
	if err != nil {
		return nil, err
	}

	return &module, nil
}

func (w *Website) GetModuleByTableName(tableName string) (*model.Module, error) {
	var module model.Module
	db := w.DB
	err := db.Where("`table_name` = ?", tableName).First(&module).Error
	if err != nil {
		return nil, err
	}

	return &module, nil
}

func (w *Website) GetModuleByUrlToken(urlToken string) (*model.Module, error) {
	var module model.Module
	db := w.DB
	err := db.Where("`url_token` = ?", urlToken).First(&module).Error
	if err != nil {
		return nil, err
	}

	return &module, nil
}

func (w *Website) SaveModule(req *request.ModuleRequest) (module *model.Module, err error) {
	if req.Id > 0 {
		module, err = w.GetModuleById(req.Id)
		if err != nil {
			return nil, err
		}
	} else {
		module = &model.Module{
			Status: 1,
		}
	}
	// 检查tableName
	exists, err := w.GetModuleByTableName(req.TableName)
	if err == nil && exists.Id != req.Id {
		return nil, errors.New(w.Lang("模型表名已存在，请更换一个"))
	}

	// 检查tableName
	exists, err = w.GetModuleByUrlToken(req.UrlToken)
	if err == nil && exists.Id != req.Id {
		return nil, errors.New(w.Lang("模型URL别名已存在，请更换一个"))
	}

	oldTableName := module.TableName
	module.TableName = req.TableName

	if oldTableName != module.TableName {
		// 表示是新表
		if w.DB.Migrator().HasTable(module.TableName) {
			return nil, errors.New(w.Lang("模型表名已存在，请更换一个"))
		}
	}
	// 检查fields
	for i := range req.Fields {
		match, err := regexp.MatchString(`^[a-z][0-9a-z_]+$`, req.Fields[i].FieldName)
		if err != nil || !match {
			return nil, errors.New(req.Fields[i].FieldName + w.Lang("命名不正确"))
		}
	}

	module.Fields = req.Fields
	module.Title = req.Title
	module.Fields = req.Fields
	module.TitleName = req.TitleName
	module.UrlToken = req.UrlToken
	module.Status = req.Status

	err = w.DB.Save(module).Error
	if err != nil {
		return
	}
	// sync table
	if oldTableName != "" && oldTableName != module.TableName {
		if w.DB.Migrator().HasTable(oldTableName) {
			w.DB.Migrator().RenameTable(oldTableName, module.TableName)
		}
	}
	module.Database = w.Mysql.Database
	tplPath := fmt.Sprintf("%s/%s", w.GetTemplateDir(), module.TableName)
	module.Migrate(w.DB, tplPath, true)

	w.DeleteCacheModules()

	return
}

func (w *Website) DeleteModuleField(moduleId uint, fieldName string) error {
	module, err := w.GetModuleById(moduleId)
	if err != nil {
		return err
	}

	if !w.DB.Migrator().HasTable(module.TableName) {
		return nil
	}

	for i, val := range module.Fields {
		if val.FieldName == fieldName {
			if module.HasColumn(w.DB, val.FieldName) {
				w.DB.Exec("ALTER TABLE ? DROP COLUMN ?", gorm.Expr(module.TableName), clause.Column{Name: val.FieldName})
			}

			module.Fields = append(module.Fields[:i], module.Fields[i+1:]...)
			break
		}
	}
	// 回写
	err = w.DB.Save(module).Error
	return err
}

func (w *Website) DeleteModule(module *model.Module) error {
	// 删除该模型的所有内容
	// 删除 archive data
	var ids []uint
	for {
		w.DB.Model(&model.Archive{}).Unscoped().Where("module_id = ?", module.Id).Limit(1000).Pluck("id", &ids)
		if len(ids) == 0 {
			break
		}
		w.DB.Unscoped().Where("id IN(?)", ids).Delete(model.ArchiveData{})
		w.DB.Unscoped().Where("id IN(?)", ids).Delete(model.Archive{})
	}
	// 删除模型表
	if w.DB.Migrator().HasTable(module.TableName) {
		w.DB.Migrator().DropTable(module.TableName)
	}
	// 删除 module
	w.DB.Delete(module)

	return nil
}

func (w *Website) DeleteCacheModules() {
	w.MemCache.Delete("modules")
}

func (w *Website) GetCacheModules() []model.Module {
	if w.DB == nil {
		return nil
	}
	var modules []model.Module

	result := w.MemCache.Get("modules")
	if result != nil {
		var ok bool
		modules, ok = result.([]model.Module)
		if ok {
			return modules
		}
	}

	w.DB.Model(model.Module{}).Where("`status` = ?", config.ContentStatusOK).Find(&modules)

	w.MemCache.Set("modules", modules, 0)

	return modules
}

func (w *Website) GetModuleFromCache(moduleId uint) *model.Module {
	modules := w.GetCacheModules()
	for i := range modules {
		if modules[i].Id == moduleId {
			return &modules[i]
		}
	}

	return nil
}

func (w *Website) GetModuleFromCacheByToken(urlToken string) *model.Module {
	modules := w.GetCacheModules()
	for i := range modules {
		if modules[i].UrlToken == urlToken {
			return &modules[i]
		}
	}

	return nil
}
