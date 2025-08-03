package repository

import (
	"context"
	"gorm.io/gorm"
	"github.com/hedeqiang/skeleton/pkg/errors"
)

// BaseRepository 基础仓储
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// DB 获取数据库实例
func (r *BaseRepository) DB() *gorm.DB {
	return r.db
}

// WithContext 创建带上下文的数据库会话
func (r *BaseRepository) WithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// Create 创建记录
func (r *BaseRepository) Create(ctx context.Context, model interface{}) error {
	if err := r.WithContext(ctx).Create(model).Error; err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to create record")
	}
	return nil
}

// Update 更新记录
func (r *BaseRepository) Update(ctx context.Context, model interface{}) error {
	if err := r.WithContext(ctx).Save(model).Error; err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to update record")
	}
	return nil
}

// Delete 删除记录
func (r *BaseRepository) Delete(ctx context.Context, model interface{}) error {
	if err := r.WithContext(ctx).Delete(model).Error; err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to delete record")
	}
	return nil
}

// FindByID 根据ID查找记录
func (r *BaseRepository) FindByID(ctx context.Context, model interface{}, id interface{}) error {
	if err := r.WithContext(ctx).First(model, id).Error; err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to find record by ID")
	}
	return nil
}

// FindOne 查找单条记录
func (r *BaseRepository) FindOne(ctx context.Context, model interface{}, query interface{}, args ...interface{}) error {
	if err := r.WithContext(ctx).Where(query, args...).First(model).Error; err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to find record")
	}
	return nil
}

// FindMany 查找多条记录
func (r *BaseRepository) FindMany(ctx context.Context, models interface{}, query interface{}, args ...interface{}) error {
	if err := r.WithContext(ctx).Where(query, args...).Find(models).Error; err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to find records")
	}
	return nil
}

// Count 统计记录数
func (r *BaseRepository) Count(ctx context.Context, model interface{}, query interface{}, args ...interface{}) (int64, error) {
	var count int64
	if err := r.WithContext(ctx).Model(model).Where(query, args...).Count(&count).Error; err != nil {
		return 0, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to count records")
	}
	return count, nil
}

// Exists 检查记录是否存在
func (r *BaseRepository) Exists(ctx context.Context, model interface{}, query interface{}, args ...interface{}) (bool, error) {
	var count int64
	if err := r.WithContext(ctx).Model(model).Where(query, args...).Count(&count).Error; err != nil {
		return false, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to check record existence")
	}
	return count > 0, nil
}