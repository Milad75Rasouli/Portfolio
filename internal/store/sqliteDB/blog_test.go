package sqlitedb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Milad75Rasouli/portfolio/internal/db"
	"github.com/Milad75Rasouli/portfolio/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBlogCRUD(t *testing.T) {
	var blogDB *StoreSqlite
	ti, err := time.Parse(timeLayout, time.Now().Format(timeLayout))
	blog := model.Blog{
		ID:         1,
		Title:      "foo bar",
		Body:       "bar baz",
		Caption:    "foo",
		ImagePath:  "/foo/bar",
		CreatedAt:  ti,
		ModifiedAt: ti,
	}

	d := SqliteInit{Folder: "data"}
	blogDB, cancel, err := d.Init(true, db.Config{}, nil)
	assert.NoError(t, err)
	defer cancel()

	{
		_, err = blogDB.CreateBlog(context.TODO(), blog)
		if err != nil {
			t.Error(err)
		}
	}

	{
		fetchedBlog, err := blogDB.GetBlogByID(context.TODO(), blog.ID)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, blog, fetchedBlog, "blog and fetchedBlog should be equal ")
		fmt.Printf("GetBlogByID: %+v\n", fetchedBlog)
	}

	{
		UpdateTest := func(base model.Blog, expected model.Blog) (result bool) {
			result = (base.Body == expected.Body) && (base.Title == expected.Title) &&
				(base.Caption == expected.Caption) && (base.ImagePath == expected.ImagePath)
			return result
		}
		updatedBlog := model.Blog{
			ID:        1,
			Title:     "bar barrrr",
			Body:      "bazzz bazzz",
			Caption:   "fooo fooo",
			ImagePath: "foo/bbbbarrr",
		}
		err = blogDB.UpdateBlogByID(context.TODO(), updatedBlog)
		assert.NoError(t, err, "unable to update blog")
		expectedBlog, err := blogDB.GetBlogByID(context.TODO(), updatedBlog.ID)
		assert.NoError(t, err, "unable to read user")
		fmt.Printf("UpdateBlogByID: %+v GetBlogByID: %+v\n", updatedBlog, expectedBlog)
		assert.True(t, UpdateTest(updatedBlog, expectedBlog), "parameters do not match")

	}

	{
		_, err = blogDB.GetBlogByID(context.TODO(), 5)
		assert.Error(t, err)

	}

	{
		err := blogDB.DeleteBlogByID(context.TODO(), blog.ID)
		assert.NoError(t, err, err)
		_, err = blogDB.GetBlogByID(context.TODO(), blog.ID)
		assert.Error(t, err)
	}

}

func TestCategoryCRUD(t *testing.T) {
	var blogDB *StoreSqlite
	category := model.Category{
		ID:   1,
		Name: "database",
	}

	d := SqliteInit{Folder: "data"}
	blogDB, cancel, err := d.Init(true, db.Config{}, nil)
	assert.NoError(t, err)
	defer cancel()

	{
		_, err = blogDB.CreateCategory(context.TODO(), category)
		if err != nil {
			t.Error(err)
		}
	}

	{
		fetchedCategory, err := blogDB.GetCategoryByID(context.TODO(), category.ID)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, category, fetchedCategory, "category and fetchCategory should be equal ")
		fmt.Printf("GetCategoryByID: %+v\n", fetchedCategory)
	}

	{
		UpdateTest := func(base model.Category, expected model.Category) (result bool) {
			result = base.Name == expected.Name
			return result
		}
		updatedCategory := model.Category{
			ID:   1,
			Name: "foo",
		}
		err = blogDB.UpdateCategoryByID(context.TODO(), updatedCategory)
		assert.NoError(t, err, "unable to update blog")
		expectedCategory, err := blogDB.GetCategoryByID(context.TODO(), updatedCategory.ID)
		assert.NoError(t, err, "unable to read user")
		fmt.Printf("UpdateCategoryByID: %+v GetCategoryByID: %+v\n", updatedCategory, expectedCategory)
		assert.True(t, UpdateTest(updatedCategory, expectedCategory), "parameters do not match")

	}

	{
		_, err = blogDB.GetCategoryByID(context.TODO(), 5)
		assert.Error(t, err)

	}

	{
		err := blogDB.DeleteBlogByID(context.TODO(), category.ID)
		assert.NoError(t, err, err)
		_, err = blogDB.GetBlogByID(context.TODO(), category.ID)
		assert.Error(t, err)
	}

}

func TestRelation(t *testing.T) {
	var blogDB *StoreSqlite
	relation := model.Relation{
		PostID:     1,
		CategoryID: 2,
	}

	relation2 := model.Relation{
		PostID:     1,
		CategoryID: 4,
	}

	d := SqliteInit{Folder: "data"}
	blogDB, cancel, err := d.Init(true, db.Config{}, nil)
	assert.NoError(t, err)
	defer cancel()

	{
		err = blogDB.CreateCategoryRelation(context.TODO(), relation)
		if err != nil {
			t.Error(err)
		}
	}
	{
		err = blogDB.CreateCategoryRelation(context.TODO(), relation2)
		if err != nil {
			t.Error(err)
		}
	}

	{
		fetchedRelation, err := blogDB.GetCategoryRelationAllByPostID(context.TODO(), relation.PostID)
		if err != nil {
			t.Error(err)
		}
		if len(fetchedRelation) > 0 {
			assert.Equal(t, relation, fetchedRelation[0], "relation and fetchedRelation should be equal ")
		} else {
			t.Error("Got less than expected items")
		}
		fmt.Printf("GetAllByPostID: %+v\n", fetchedRelation)
	}

	{
		fetchedRelation, err := blogDB.GetCategoryRelationAllByCategoryID(context.TODO(), relation.CategoryID)
		if err != nil {
			t.Error(err)
		}
		if len(fetchedRelation) > 0 {
			assert.Equal(t, relation, fetchedRelation[0], "relation and fetchedRelation should be equal ")
		} else {
			t.Error("Got less than expected items")
		}
		fmt.Printf("GetAllByCategoryID: %+v\n", fetchedRelation)
	}

	{
		_, err = blogDB.GetCategoryRelationAllByCategoryID(context.TODO(), 5)
		assert.Error(t, err)

	}

	{
		var fetchedRelation []model.Relation
		err := blogDB.DeleteCategoryRelationAllByCategoryID(context.TODO(), relation.CategoryID)
		assert.NoError(t, err, err)
		fetchedRelation, err = blogDB.GetCategoryRelationAllByCategoryID(context.TODO(), relation.CategoryID)
		assert.Error(t, err)
		fmt.Printf("DeleteAllByCategoryID:GetAllByCategoryID: %+v\n", fetchedRelation)
	}

	{
		var fetchedRelation []model.Relation
		err := blogDB.DeleteCategoryRelationAllByPostID(context.TODO(), relation2.PostID)
		assert.NoError(t, err, err)
		fetchedRelation, err = blogDB.GetCategoryRelationAllByPostID(context.TODO(), relation2.PostID)
		assert.Error(t, err)
		fmt.Printf("DeleteAllByPostID:GetAllByPostID: %+v\n", fetchedRelation)
	}
}

func TestGeneral(t *testing.T) {
	var blogDB *StoreSqlite

	d := SqliteInit{Folder: "data"}
	blogDB, cancel, err := d.Init(true, db.Config{}, nil)
	assert.NoError(t, err)
	defer cancel()

	{
		ti, err := time.Parse(timeLayout, time.Now().Format(timeLayout))
		assert.NoError(t, err)

		blogs := [...]model.Blog{
			{
				Title:      "foo in action",
				Body:       "bar bar bar",
				Caption:    "baz baz",
				CreatedAt:  ti,
				ModifiedAt: ti,
			},
			{
				Title:      "bar in action",
				Body:       "baz bar bar",
				Caption:    "bar baz",
				CreatedAt:  ti,
				ModifiedAt: ti,
			},
			{
				Title:      "baz in action",
				Body:       "foo bar bar",
				Caption:    "bar baz",
				CreatedAt:  ti,
				ModifiedAt: ti,
			},
		}
		for i := 0; i < len(blogs); i++ {
			id, err := blogDB.CreateBlog(context.TODO(), blogs[i])
			assert.NoError(t, err, blogs[i])
			blogs[i].ID = id
		}

		category := [...]model.Category{
			{
				Name: "database",
			}, {
				Name: "redis",
			}, {
				Name: "message broker",
			}, {
				Name: "algorithm",
			}, {
				Name: "postgresql",
			},
		}
		for i := 0; i < len(category); i++ {
			id, err := blogDB.CreateCategory(context.TODO(), category[i])
			assert.NoError(t, err, category[i])
			category[i].ID = id
		}

		categoryRelation := []model.Relation{
			{PostID: 1, CategoryID: 2},
			{PostID: 1, CategoryID: 1},
			{PostID: 1, CategoryID: 3},
			{PostID: 2, CategoryID: 4},
			{PostID: 2, CategoryID: 5},
			{PostID: 2, CategoryID: 1},
			{PostID: 3, CategoryID: 1},
			{PostID: 3, CategoryID: 5},
		}
		for i := 0; i < len(categoryRelation); i++ {
			err := blogDB.CreateCategoryRelation(context.TODO(), categoryRelation[i])
			assert.NoError(t, err, categoryRelation[i])
		}
		// err = blogDB.DeleteBlogByID(context.TODO(), 1)
		// assert.NoError(t, err)
		err = blogDB.DeleteCategoryRelationAllByPostID(context.TODO(), 1)
		assert.NoError(t, err)

		postWithCategory, err := blogDB.GetAllPostsWithCategory(context.Background())
		assert.NoError(t, err)
		fmt.Printf("TestGeneral:GetAllPostsWithCategory %+v\n", postWithCategory)
	}

}
