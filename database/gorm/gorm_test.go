package gorm

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	_ "gorm.io/driver/postgres"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/file"
)

type User struct {
	orm.Model
	orm.SoftDeletes
	Name    string
	Avatar  string
	Address *Address
	Books   []*Book
	House   *House   `gorm:"polymorphic:Houseable"`
	Phones  []*Phone `gorm:"polymorphic:Phoneable"`
	Roles   []*Role  `gorm:"many2many:role_user"`
}

type Role struct {
	orm.Model
	Name  string
	Users []*User `gorm:"many2many:role_user"`
}

type Address struct {
	orm.Model
	UserID   uint
	Name     string
	Province string
	User     *User
}

type Book struct {
	orm.Model
	UserID uint
	Name   string
	User   *User
	Author *Author
}

type Author struct {
	orm.Model
	BookID uint
	Name   string
}

type House struct {
	orm.Model
	Name          string
	HouseableID   uint
	HouseableType string
}

type Phone struct {
	orm.Model
	Name          string
	PhoneableID   uint
	PhoneableType string
}

type GormQueryTestSuite struct {
	suite.Suite
	queries map[contractsorm.Driver]contractsorm.Query
}

func TestGormQueryTestSuite(t *testing.T) {
	mysqlPool, mysqlResource, mysqlDB, err := MysqlDocker()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}

	postgresqlPool, postgresqlResource, postgresqlDB, err := PostgresqlDocker()
	if err != nil {
		log.Fatalf("Init postgresql error: %s", err)
	}

	_, _, sqliteDB, err := SqliteDocker(dbDatabase)
	if err != nil {
		log.Fatalf("Init sqlite error: %s", err)
	}

	sqlserverPool, sqlserverResource, sqlserverDB, err := SqlserverDocker()
	if err != nil {
		log.Fatalf("Init sqlserver error: %s", err)
	}

	suite.Run(t, &GormQueryTestSuite{
		queries: map[contractsorm.Driver]contractsorm.Query{
			contractsorm.DriverMysql:      mysqlDB,
			contractsorm.DriverPostgresql: postgresqlDB,
			contractsorm.DriverSqlite:     sqliteDB,
			contractsorm.DriverSqlserver:  sqlserverDB,
		},
	})

	file.Remove(dbDatabase)

	if err := mysqlPool.Purge(mysqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := postgresqlPool.Purge(postgresqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := sqlserverPool.Purge(sqlserverResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func (s *GormQueryTestSuite) SetupTest() {
}

func (s *GormQueryTestSuite) TestAssociation() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			tests := []struct {
				description string
				setup       func()
			}{
				{
					description: "Find",
					setup: func() {
						user := &User{
							Name: "association_find_name",
							Address: &Address{
								Name: "association_find_address",
							},
						}

						s.Nil(query.Select(orm.Associations).Create(&user))
						s.True(user.ID > 0)
						s.True(user.Address.ID > 0)

						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)

						var userAddress Address
						s.Nil(query.Model(&user1).Association("Address").Find(&userAddress))
						s.True(userAddress.ID > 0)
						s.Equal("association_find_address", userAddress.Name)
					},
				},
				{
					description: "hasOne Append",
					setup: func() {
						user := &User{
							Name: "association_has_one_append_name",
							Address: &Address{
								Name: "association_has_one_append_address",
							},
						}

						s.Nil(query.Select(orm.Associations).Create(&user))
						s.True(user.ID > 0)
						s.True(user.Address.ID > 0)

						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(query.Model(&user1).Association("Address").Append(&Address{Name: "association_has_one_append_address1"}))

						s.Nil(query.Load(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.Equal("association_has_one_append_address1", user1.Address.Name)
					},
				},
				{
					description: "hasMany Append",
					setup: func() {
						user := &User{
							Name: "association_has_many_append_name",
							Books: []*Book{
								{Name: "association_has_many_append_address1"},
								{Name: "association_has_many_append_address2"},
							},
						}

						s.Nil(query.Select(orm.Associations).Create(&user))
						s.True(user.ID > 0)
						s.True(user.Books[0].ID > 0)
						s.True(user.Books[1].ID > 0)

						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(query.Model(&user1).Association("Books").Append(&Book{Name: "association_has_many_append_address3"}))

						s.Nil(query.Load(&user1, "Books"))
						s.Equal(3, len(user1.Books))
						s.Equal("association_has_many_append_address3", user1.Books[2].Name)
					},
				},
				{
					description: "hasOne Replace",
					setup: func() {
						user := &User{
							Name: "association_has_one_append_name",
							Address: &Address{
								Name: "association_has_one_append_address",
							},
						}

						s.Nil(query.Select(orm.Associations).Create(&user))
						s.True(user.ID > 0)
						s.True(user.Address.ID > 0)

						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(query.Model(&user1).Association("Address").Replace(&Address{Name: "association_has_one_append_address1"}))

						s.Nil(query.Load(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.Equal("association_has_one_append_address1", user1.Address.Name)
					},
				},
				{
					description: "hasMany Replace",
					setup: func() {
						user := &User{
							Name: "association_has_many_replace_name",
							Books: []*Book{
								{Name: "association_has_many_replace_address1"},
								{Name: "association_has_many_replace_address2"},
							},
						}

						s.Nil(query.Select(orm.Associations).Create(&user))
						s.True(user.ID > 0)
						s.True(user.Books[0].ID > 0)
						s.True(user.Books[1].ID > 0)

						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(query.Model(&user1).Association("Books").Replace(&Book{Name: "association_has_many_replace_address3"}))

						s.Nil(query.Load(&user1, "Books"))
						s.Equal(1, len(user1.Books))
						s.Equal("association_has_many_replace_address3", user1.Books[0].Name)
					},
				},
				{
					description: "Delete",
					setup: func() {
						user := &User{
							Name: "association_delete_name",
							Address: &Address{
								Name: "association_delete_address",
							},
						}

						s.Nil(query.Select(orm.Associations).Create(&user))
						s.True(user.ID > 0)
						s.True(user.Address.ID > 0)

						// No ID when Delete
						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(query.Model(&user1).Association("Address").Delete(&Address{Name: "association_delete_address"}))

						s.Nil(query.Load(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.Equal("association_delete_address", user1.Address.Name)

						// Has ID when Delete
						var user2 User
						s.Nil(query.Find(&user2, user.ID))
						s.True(user2.ID > 0)
						var userAddress Address
						userAddress.ID = user1.Address.ID
						s.Nil(query.Model(&user2).Association("Address").Delete(&userAddress))

						s.Nil(query.Load(&user2, "Address"))
						s.Nil(user2.Address)
					},
				},
				{
					description: "Clear",
					setup: func() {
						user := &User{
							Name: "association_clear_name",
							Address: &Address{
								Name: "association_clear_address",
							},
						}

						s.Nil(query.Select(orm.Associations).Create(&user))
						s.True(user.ID > 0)
						s.True(user.Address.ID > 0)

						// No ID when Delete
						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(query.Model(&user1).Association("Address").Clear())

						s.Nil(query.Load(&user1, "Address"))
						s.Nil(user1.Address)
					},
				},
				{
					description: "Count",
					setup: func() {
						user := &User{
							Name: "association_count_name",
							Books: []*Book{
								{Name: "association_count_address1"},
								{Name: "association_count_address2"},
							},
						}

						s.Nil(query.Select(orm.Associations).Create(&user))
						s.True(user.ID > 0)
						s.True(user.Books[0].ID > 0)
						s.True(user.Books[1].ID > 0)

						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Equal(int64(2), query.Model(&user1).Association("Books").Count())
					},
				},
			}

			for _, test := range tests {
				s.Run(test.description, func() {
					test.setup()
				})
			}
		})
	}
}

func (s *GormQueryTestSuite) TestCount() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "count_user", Avatar: "count_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "count_user", Avatar: "count_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var count int64
			s.Nil(query.Model(&User{}).Where("name = ?", "count_user").Count(&count))
			s.True(count > 0)

			var count1 int64
			s.Nil(query.Table("users").Where("name = ?", "count_user").Count(&count1))
			s.True(count1 > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestCreate() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			tests := []struct {
				description string
				setup       func(description string)
			}{
				{
					description: "success when create with no relationships",
					setup: func(description string) {
						user := User{Name: "create_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
						user.Address.Name = "create_address"
						user.Books[0].Name = "create_book0"
						user.Books[1].Name = "create_book1"
						s.Nil(query.Create(&user), description)
						s.True(user.ID > 0, description)
						s.True(user.Address.ID == 0, description)
						s.True(user.Books[0].ID == 0, description)
						s.True(user.Books[1].ID == 0, description)
					},
				},
				{
					description: "success when create with select orm.Associations",
					setup: func(description string) {
						user := User{Name: "create_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
						user.Address.Name = "create_address"
						user.Books[0].Name = "create_book0"
						user.Books[1].Name = "create_book1"
						s.Nil(query.Select(orm.Associations).Create(&user), description)
						s.True(user.ID > 0, description)
						s.True(user.Address.ID > 0, description)
						s.True(user.Books[0].ID > 0, description)
						s.True(user.Books[1].ID > 0, description)
					},
				},
				{
					description: "success when create with select fields",
					setup: func(description string) {
						user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
						user.Address.Name = "create_address"
						user.Books[0].Name = "create_book0"
						user.Books[1].Name = "create_book1"
						s.Nil(query.Select("Name", "Avatar", "Address").Create(&user), description)
						s.True(user.ID > 0, description)
						s.True(user.Address.ID > 0, description)
						s.True(user.Books[0].ID == 0, description)
						s.True(user.Books[1].ID == 0, description)
					},
				},
				{
					description: "success when create with omit fields",
					setup: func(description string) {
						user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
						user.Address.Name = "create_address"
						user.Books[0].Name = "create_book0"
						user.Books[1].Name = "create_book1"
						s.Nil(query.Omit("Address").Create(&user), description)
						s.True(user.ID > 0, description)
						s.True(user.Address.ID == 0, description)
						s.True(user.Books[0].ID > 0, description)
						s.True(user.Books[1].ID > 0, description)
					},
				},
				{
					description: "success create with omit orm.Associations",
					setup: func(description string) {
						user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
						user.Address.Name = "create_address"
						user.Books[0].Name = "create_book0"
						user.Books[1].Name = "create_book1"
						s.Nil(query.Omit(orm.Associations).Create(&user), description)
						s.True(user.ID > 0, description)
						s.True(user.Address.ID == 0, description)
						s.True(user.Books[0].ID == 0, description)
						s.True(user.Books[1].ID == 0, description)
					},
				},
				{
					description: "error when set select and omit at the same time",
					setup: func(description string) {
						user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
						user.Address.Name = "create_address"
						user.Books[0].Name = "create_book0"
						user.Books[1].Name = "create_book1"
						s.EqualError(query.Omit(orm.Associations).Select("Name").Create(&user), "cannot set Select and Omits at the same time", description)
					},
				},
				{
					description: "error when select that set fields and orm.Associations at the same time",
					setup: func(description string) {
						user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
						user.Address.Name = "create_address"
						user.Books[0].Name = "create_book0"
						user.Books[1].Name = "create_book1"
						s.EqualError(query.Select("Name", orm.Associations).Create(&user), "cannot set orm.Associations and other fields at the same time", description)
					},
				},
				{
					description: "error when omit that set fields and orm.Associations at the same time",
					setup: func(description string) {
						user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
						user.Address.Name = "create_address"
						user.Books[0].Name = "create_book0"
						user.Books[1].Name = "create_book1"
						s.EqualError(query.Omit("Name", orm.Associations).Create(&user), "cannot set orm.Associations and other fields at the same time", description)
					},
				},
			}
			for _, test := range tests {
				test.setup(test.description)
			}
		})
	}
}

func (s *GormQueryTestSuite) TestDelete() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "delete_user", Avatar: "delete_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			res, err := query.Delete(&user)
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user1 User
			s.Nil(query.Find(&user1, user.ID))
			s.Equal(uint(0), user1.ID)

			user2 := User{Name: "delete_user", Avatar: "delete_avatar"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			res, err = query.Delete(&User{}, user2.ID)
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user3 User
			s.Nil(query.Find(&user3, user2.ID))
			s.Equal(uint(0), user3.ID)

			users := []User{{Name: "delete_user", Avatar: "delete_avatar"}, {Name: "delete_user1", Avatar: "delete_avatar1"}}
			s.Nil(query.Create(&users))
			s.True(users[0].ID > 0)
			s.True(users[1].ID > 0)

			res, err = query.Delete(&User{}, []uint{users[0].ID, users[1].ID})
			s.Equal(int64(2), res.RowsAffected)
			s.Nil(err)

			var count int64
			s.Nil(query.Model(&User{}).Where("name", "delete_user").OrWhere("name", "delete_user1").Count(&count))
			s.True(count == 0)
		})
	}
}

func (s *GormQueryTestSuite) TestDistinct() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "distinct_user", Avatar: "distinct_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "distinct_user", Avatar: "distinct_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Distinct("name").Find(&users, []uint{user.ID, user1.ID}))
			s.Equal(1, len(users))
		})
	}
}

func (s *GormQueryTestSuite) TestExec() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			res, err := query.Exec("INSERT INTO users (name, avatar, created_at, updated_at) VALUES ('exec_user', 'exec_avatar', '2023-03-09 18:56:33', '2023-03-09 18:56:35');")
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user User
			err = query.Where("name", "exec_user").First(&user)
			s.Nil(err)
			s.True(user.ID > 0)

			res, err = query.Exec(fmt.Sprintf("UPDATE users set name = 'exec_user1' where id = %d", user.ID))
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			res, err = query.Exec(fmt.Sprintf("DELETE FROM users where id = %d", user.ID))
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)
		})
	}
}

func (s *GormQueryTestSuite) TestFind() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "find_user"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			var user2 User
			s.Nil(query.Find(&user2, user.ID))
			s.True(user2.ID > 0)

			var user3 []User
			s.Nil(query.Find(&user3, []uint{user.ID}))
			s.Equal(1, len(user3))

			var user4 []User
			s.Nil(query.Where("id in ?", []uint{user.ID}).Find(&user4))
			s.Equal(1, len(user4))
		})
	}
}

func (s *GormQueryTestSuite) TestFirst() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "first_user"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			var user1 User
			s.Nil(query.Where("name", "first_user").First(&user1))
			s.True(user1.ID > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestFirstOrCreate() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			var user User
			s.Nil(query.Where("avatar", "first_or_create_avatar").FirstOrCreate(&user, User{Name: "first_or_create_user"}))
			s.True(user.ID > 0)

			var user1 User
			s.Nil(query.Where("avatar", "first_or_create_avatar").FirstOrCreate(&user1, User{Name: "user"}, User{Avatar: "first_or_create_avatar1"}))
			s.True(user1.ID > 0)
			s.True(user1.Avatar == "first_or_create_avatar1")
		})
	}
}

func (s *GormQueryTestSuite) TestFirstOr() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			var user User
			s.Nil(query.Where("name", "first_or_user").FirstOr(&user, func() error {
				user.Name = "goravel"

				return nil
			}))
			s.Equal(uint(0), user.ID)
			s.Equal("goravel", user.Name)

			var user1 User
			s.EqualError(query.Where("name", "first_or_user").FirstOr(&user1, func() error {
				return errors.New("error")
			}), "error")
			s.Equal(uint(0), user1.ID)
		})
	}
}

func (s *GormQueryTestSuite) TestFirstOrFail() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			var user User
			s.Equal(orm.ErrRecordNotFound, query.Where("name", "first_or_fail_user").FirstOrFail(&user))
			s.Equal(uint(0), user.ID)
		})
	}
}

func (s *GormQueryTestSuite) TestFirstOrNew() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			var user User
			s.Nil(query.FirstOrNew(&user, User{Name: "first_or_new_name"}))
			s.Equal(uint(0), user.ID)
			s.Equal("first_or_new_name", user.Name)
			s.Equal("", user.Avatar)

			var user1 User
			s.Nil(query.FirstOrNew(&user1, User{Name: "first_or_new_name"}, User{Avatar: "first_or_new_avatar"}))
			s.Equal(uint(0), user1.ID)
			s.Equal("first_or_new_name", user1.Name)
			s.Equal("first_or_new_avatar", user1.Avatar)
		})
	}
}

func (s *GormQueryTestSuite) TestGet() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "get_user"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			var user5 []User
			s.Nil(query.Where("id in ?", []uint{user.ID}).Get(&user5))
			s.Equal(1, len(user5))
		})
	}
}

func (s *GormQueryTestSuite) TestJoin() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "join_user", Avatar: "join_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			userAddress := Address{UserID: user.ID, Name: "join_address", Province: "join_province"}
			s.Nil(query.Create(&userAddress))
			s.True(userAddress.ID > 0)

			type Result struct {
				UserName        string
				UserAddressName string
			}
			var result []Result
			s.Nil(query.Model(&User{}).Where("users.id = ?", user.ID).Join("left join addresses ua on users.id = ua.user_id").
				Select("users.name user_name, ua.name user_address_name").Get(&result))
			s.Equal(1, len(result))
			s.Equal("join_user", result[0].UserName)
			s.Equal("join_address", result[0].UserAddressName)
		})
	}
}

func (s *GormQueryTestSuite) TestOffset() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "offset_user", Avatar: "offset_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "offset_user", Avatar: "offset_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Where("name = ?", "offset_user").Offset(1).Limit(1).Get(&user2))
			s.True(len(user2) > 0)
			s.True(user2[0].ID > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestOrder() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "order_user", Avatar: "order_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "order_user", Avatar: "order_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Where("name = ?", "order_user").Order("id desc").Order("name asc").Get(&user2))
			s.True(len(user2) > 0)
			s.True(user2[0].ID > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestPaginate() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "paginate_user", Avatar: "paginate_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "paginate_user", Avatar: "paginate_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "paginate_user", Avatar: "paginate_avatar2"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			user3 := User{Name: "paginate_user", Avatar: "paginate_avatar3"}
			s.Nil(query.Create(&user3))
			s.True(user3.ID > 0)

			var users []User
			var total int64
			s.Nil(query.Where("name = ?", "paginate_user").Paginate(1, 3, &users, nil))
			s.Equal(3, len(users))

			s.Nil(query.Where("name = ?", "paginate_user").Paginate(2, 3, &users, &total))
			s.Equal(1, len(users))
			s.Equal(int64(4), total)

			s.Nil(query.Model(User{}).Where("name = ?", "paginate_user").Paginate(1, 3, &users, &total))
			s.Equal(3, len(users))
			s.Equal(int64(4), total)

			s.Nil(query.Table("users").Where("name = ?", "paginate_user").Paginate(1, 3, &users, &total))
			s.Equal(3, len(users))
			s.Equal(int64(4), total)
		})
	}
}

func (s *GormQueryTestSuite) TestPluck() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "pluck_user", Avatar: "pluck_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "pluck_user", Avatar: "pluck_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var avatars []string
			s.Nil(query.Model(&User{}).Where("name = ?", "pluck_user").Pluck("avatar", &avatars))
			s.Equal(2, len(avatars))
			s.Equal("pluck_avatar", avatars[0])
			s.Equal("pluck_avatar1", avatars[1])
		})
	}
}

func (s *GormQueryTestSuite) TestHasOne() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "has_one_name",
				Address: &Address{
					Name: "has_one_address",
				},
			}

			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)

			var user1 User
			s.Nil(query.With("Address").Where("name = ?", "has_one_name").First(&user1))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestHasOneMorph() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "has_one_morph_name",
				House: &House{
					Name: "has_one_morph_house",
				},
			}
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.House.ID > 0)

			var user1 User
			s.Nil(query.With("House").Where("name = ?", "has_one_morph_name").First(&user1))
			s.True(user.ID > 0)
			s.True(user.Name == "has_one_morph_name")
			s.True(user.House.ID > 0)
			s.True(user.House.Name == "has_one_morph_house")

			var house House
			s.Nil(query.Where("name = ?", "has_one_morph_house").Where("houseable_type = ?", "users").Where("houseable_id = ?", user.ID).First(&house))
			s.True(house.ID > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestHasMany() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "has_many_name",
				Books: []*Book{
					{Name: "has_many_book1"},
					{Name: "has_many_book2"},
				},
			}

			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[1].ID > 0)

			var user1 User
			s.Nil(query.With("Books").Where("name = ?", "has_many_name").First(&user1))
			s.True(user.ID > 0)
			s.True(len(user.Books) == 2)
		})
	}
}

func (s *GormQueryTestSuite) TestHasManyMorph() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "has_many_morph_name",
				Phones: []*Phone{
					{Name: "has_many_morph_phone1"},
					{Name: "has_many_morph_phone2"},
				},
			}
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Phones[0].ID > 0)
			s.True(user.Phones[1].ID > 0)

			var user1 User
			s.Nil(query.With("Phones").Where("name = ?", "has_many_morph_name").First(&user1))
			s.True(user.ID > 0)
			s.True(user.Name == "has_many_morph_name")
			s.True(len(user.Phones) == 2)
			s.True(user.Phones[0].Name == "has_many_morph_phone1")
			s.True(user.Phones[1].Name == "has_many_morph_phone2")

			var phones []Phone
			s.Nil(query.Where("name like ?", "has_many_morph_phone%").Where("phoneable_type = ?", "users").Where("phoneable_id = ?", user.ID).Find(&phones))
			s.True(len(phones) == 2)
		})
	}
}

func (s *GormQueryTestSuite) TestBelongsTo() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "belongs_to_name",
				Address: &Address{
					Name: "belongs_to_address",
				},
			}

			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)

			var userAddress Address
			s.Nil(query.With("User").Where("name = ?", "belongs_to_address").First(&userAddress))
			s.True(userAddress.ID > 0)
			s.True(userAddress.User.ID > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestManyToMany() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "many_to_many_name",
				Roles: []*Role{
					{Name: "many_to_many_role1"},
					{Name: "many_to_many_role2"},
				},
			}

			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Roles[0].ID > 0)
			s.True(user.Roles[1].ID > 0)

			var user1 User
			s.Nil(query.With("Roles").Where("name = ?", "many_to_many_name").First(&user1))
			s.True(user.ID > 0)
			s.True(len(user.Roles) == 2)

			var role Role
			s.Nil(query.With("Users").Where("name = ?", "many_to_many_role1").First(&role))
			s.True(role.ID > 0)
			s.True(len(role.Users) == 1)
			s.Equal("many_to_many_name", role.Users[0].Name)
		})
	}
}

func (s *GormQueryTestSuite) TestLimit() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "limit_user", Avatar: "limit_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "limit_user", Avatar: "limit_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Where("name = ?", "limit_user").Limit(1).Get(&user2))
			s.True(len(user2) > 0)
			s.True(user2[0].ID > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestLoad() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "load_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
			user.Address.Name = "load_address"
			user.Books[0].Name = "load_book0"
			user.Books[1].Name = "load_book1"
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[1].ID > 0)

			tests := []struct {
				description string
				setup       func(description string)
			}{
				{
					description: "simple load relationship",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.True(len(user1.Books) == 0)
						s.Nil(query.Load(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.True(len(user1.Books) == 0)
						s.Nil(query.Load(&user1, "Books"))
						s.True(user1.Address.ID > 0)
						s.True(len(user1.Books) == 2)
					},
				},
				{
					description: "load relationship with simple condition",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.Equal(0, len(user1.Books))
						s.Nil(query.Load(&user1, "Books", "name = ?", "load_book0"))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.Equal(1, len(user1.Books))
						s.Equal("load_book0", user.Books[0].Name)
					},
				},
				{
					description: "load relationship with func condition",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.Equal(0, len(user1.Books))
						s.Nil(query.Load(&user1, "Books", func(query contractsorm.Query) contractsorm.Query {
							return query.Where("name = ?", "load_book0")
						}))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.Equal(1, len(user1.Books))
						s.Equal("load_book0", user.Books[0].Name)
					},
				},
				{
					description: "error when relation is empty",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.Equal(0, len(user1.Books))
						s.EqualError(query.Load(&user1, ""), "relation cannot be empty")
					},
				},
				{
					description: "error when id is nil",
					setup: func(description string) {
						type UserNoID struct {
							Name   string
							Avatar string
						}
						var userNoID UserNoID
						s.EqualError(query.Load(&userNoID, "Book"), "id cannot be empty")
					},
				},
			}
			for _, test := range tests {
				test.setup(test.description)
			}
		})
	}
}

func (s *GormQueryTestSuite) TestLoadMissing() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "load_missing_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
			user.Address.Name = "load_missing_address"
			user.Books[0].Name = "load_missing_book0"
			user.Books[1].Name = "load_missing_book1"
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[1].ID > 0)

			tests := []struct {
				description string
				setup       func(description string)
			}{
				{
					description: "load when missing",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.True(len(user1.Books) == 0)
						s.Nil(query.LoadMissing(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.True(len(user1.Books) == 0)
						s.Nil(query.LoadMissing(&user1, "Books"))
						s.True(user1.Address.ID > 0)
						s.True(len(user1.Books) == 2)
					},
				},
				{
					description: "don't load when not missing",
					setup: func(description string) {
						var user1 User
						s.Nil(query.With("Books", "name = ?", "load_missing_book0").Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.True(len(user1.Books) == 1)
						s.Nil(query.LoadMissing(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.Nil(query.LoadMissing(&user1, "Books"))
						s.True(len(user1.Books) == 1)
					},
				},
			}
			for _, test := range tests {
				test.setup(test.description)
			}
		})
	}
}

func (s *GormQueryTestSuite) TestRaw() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "raw_user", Avatar: "raw_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			var user1 User
			s.Nil(query.Raw("SELECT id, name FROM users WHERE name = ?", "raw_user").Scan(&user1))
			s.True(user1.ID > 0)
			s.Equal("raw_user", user1.Name)
			s.Equal("", user1.Avatar)
		})
	}
}

func (s *GormQueryTestSuite) TestScope() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			users := []User{{Name: "scope_user", Avatar: "scope_avatar"}, {Name: "scope_user1", Avatar: "scope_avatar1"}}
			s.Nil(query.Create(&users))
			s.True(users[0].ID > 0)
			s.True(users[1].ID > 0)

			var users1 []User
			s.Nil(query.Scopes(paginator("1", "1")).Find(&users1))

			s.Equal(1, len(users1))
			s.True(users1[0].ID > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestSelect() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "select_user", Avatar: "select_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "select_user", Avatar: "select_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "select_user1", Avatar: "select_avatar1"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			type Result struct {
				Name  string
				Count string
			}
			var result []Result
			s.Nil(query.Model(&User{}).Select("name, count(avatar) as count").Where("id in ?", []uint{user.ID, user1.ID, user2.ID}).Group("name").Get(&result))
			s.Equal(2, len(result))
			s.Equal("select_user", result[0].Name)
			s.Equal("2", result[0].Count)
			s.Equal("select_user1", result[1].Name)
			s.Equal("1", result[1].Count)

			var result1 []Result
			s.Nil(query.Model(&User{}).Select("name, count(avatar) as count").Group("name").Having("name = ?", "select_user").Get(&result1))

			s.Equal(1, len(result1))
			s.Equal("select_user", result1[0].Name)
			s.Equal("2", result1[0].Count)
		})
	}
}

func (s *GormQueryTestSuite) TestSoftDelete() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "soft_delete_user", Avatar: "soft_delete_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			res, err := query.Where("name = ?", "soft_delete_user").Delete(&User{})
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user1 User
			s.Nil(query.Find(&user1, user.ID))
			s.Equal(uint(0), user1.ID)

			var user2 User
			s.Nil(query.WithTrashed().Find(&user2, user.ID))
			s.True(user2.ID > 0)

			res, err = query.Where("name = ?", "soft_delete_user").ForceDelete(&User{})
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user3 User
			s.Nil(query.WithTrashed().Find(&user3, user.ID))
			s.Equal(uint(0), user3.ID)
		})
	}
}

func (s *GormQueryTestSuite) TestTransactionSuccess() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "transaction_success_user", Avatar: "transaction_success_avatar"}
			user1 := User{Name: "transaction_success_user1", Avatar: "transaction_success_avatar1"}
			tx, err := query.Begin()
			s.Nil(err)
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))
			s.Nil(tx.Commit())

			var user2, user3 User
			s.Nil(query.Find(&user2, user.ID))
			s.Nil(query.Find(&user3, user1.ID))
		})
	}
}

func (s *GormQueryTestSuite) TestTransactionError() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "transaction_error_user", Avatar: "transaction_error_avatar"}
			user1 := User{Name: "transaction_error_user1", Avatar: "transaction_error_avatar1"}
			tx, err := query.Begin()
			s.Nil(err)
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))
			s.Nil(tx.Rollback())

			var users []User
			s.Nil(query.Where("name = ? or name = ?", "transaction_error_user", "transaction_error_user1").Find(&users))
			s.Equal(0, len(users))
		})
	}
}

func (s *GormQueryTestSuite) TestUpdate() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "update_user", Avatar: "update_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user.Name = "update_user1"
			s.Nil(query.Save(&user))
			s.Nil(query.Model(&User{}).Where("id = ?", user.ID).Update("avatar", "update_avatar1"))

			var user1 User
			s.Nil(query.Find(&user1, user.ID))
			s.Equal("update_user1", user1.Name)
			s.Equal("update_avatar1", user1.Avatar)
		})
	}
}

func (s *GormQueryTestSuite) TestUpdates() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			users := []User{{Name: "updates_user", Avatar: "updates_avatar"}, {Name: "updates_user", Avatar: "updates_avatar1"}}
			s.Nil(query.Create(&users))
			s.True(users[0].ID > 0)
			s.True(users[1].ID > 0)

			res, err := query.Where("name", "updates_user").Updates(User{Avatar: "updates_avatar2"})
			s.Equal(int64(2), res.RowsAffected)
			s.Nil(err)

			var count int64
			err = query.Model(User{}).Where("avatar", "updates_avatar2").Count(&count)
			s.Equal(int64(2), count)
			s.Nil(err)
		})
	}
}

func (s *GormQueryTestSuite) TestUpdateOrCreate() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			var user User
			err := query.UpdateOrCreate(&user, User{Name: "update_or_create_user"}, User{Avatar: "update_or_create_avatar"})
			s.Nil(err)
			s.True(user.ID > 0)

			var user1 User
			err = query.Where("name", "update_or_create_user").Find(&user1)
			s.Nil(err)
			s.True(user1.ID > 0)

			var user2 User
			err = query.Where("name", "update_or_create_user").UpdateOrCreate(&user2, User{Name: "update_or_create_user"}, User{Avatar: "update_or_create_avatar1"})
			s.Nil(err)
			s.True(user2.ID > 0)
			s.Equal("update_or_create_avatar1", user2.Avatar)

			var user3 User
			err = query.Where("avatar", "update_or_create_avatar1").Find(&user3)
			s.Nil(err)
			s.True(user3.ID > 0)

			var count int64
			err = query.Model(User{}).Where("name", "update_or_create_user").Count(&count)
			s.Nil(err)
			s.Equal(int64(1), count)
		})
	}
}

func (s *GormQueryTestSuite) TestWhere() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "where_user", Avatar: "where_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_user1", Avatar: "where_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Where("name = ?", "where_user").OrWhere("avatar = ?", "where_avatar1").Find(&user2))
			s.True(len(user2) > 0)

			var user3 User
			s.Nil(query.Where("name = 'where_user'").Find(&user3))
			s.True(user3.ID > 0)

			var user4 User
			s.Nil(query.Where("name", "where_user").Find(&user4))
			s.True(user4.ID > 0)
		})
	}
}

func (s *GormQueryTestSuite) TestWith() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "with_user", Address: &Address{
				Name: "with_address",
			}, Books: []*Book{{
				Name: "with_book0",
			}, {
				Name: "with_book1",
			}}}
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[1].ID > 0)

			tests := []struct {
				description string
				setup       func(description string)
			}{
				{
					description: "simple",
					setup: func(description string) {
						var user1 User
						s.Nil(query.With("Address").With("Books").Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.True(user1.Address.ID > 0)
						s.True(user1.Books[0].ID > 0)
						s.True(user1.Books[1].ID > 0)
					},
				},
				{
					description: "with simple conditions",
					setup: func(description string) {
						var user1 User
						s.Nil(query.With("Books", "name = ?", "with_book0").Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.Equal(1, len(user1.Books))
						s.Equal("with_book0", user1.Books[0].Name)
					},
				},
				{
					description: "with func conditions",
					setup: func(description string) {
						var user1 User
						s.Nil(query.With("Books", func(query contractsorm.Query) contractsorm.Query {
							return query.Where("name = ?", "with_book0")
						}).Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.Equal(1, len(user1.Books))
						s.Equal("with_book0", user1.Books[0].Name)
					},
				},
			}
			for _, test := range tests {
				test.setup(test.description)
			}
		})
	}
}

func (s *GormQueryTestSuite) TestWithNesting() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "with_nesting_user", Books: []*Book{{
				Name:   "with_nesting_book0",
				Author: &Author{Name: "with_nesting_author0"},
			}, {
				Name:   "with_nesting_book1",
				Author: &Author{Name: "with_nesting_author1"},
			}}}
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[0].Author.ID > 0)
			s.True(user.Books[1].ID > 0)
			s.True(user.Books[1].Author.ID > 0)

			var user1 User
			s.Nil(query.With("Books.Author").Find(&user1, user.ID))
			s.True(user1.ID > 0)
			s.Equal("with_nesting_user", user1.Name)
			s.True(user1.Books[0].ID > 0)
			s.Equal("with_nesting_book0", user1.Books[0].Name)
			s.True(user1.Books[0].Author.ID > 0)
			s.Equal("with_nesting_author0", user1.Books[0].Author.Name)
			s.True(user1.Books[1].ID > 0)
			s.Equal("with_nesting_book1", user1.Books[1].Name)
			s.True(user1.Books[1].Author.ID > 0)
			s.Equal("with_nesting_author1", user1.Books[1].Author.Name)
		})
	}
}

func TestReadWriteSeparate(t *testing.T) {
	readMysqlPool, readMysqlResource, readMysqlDB, err := MysqlDocker()
	if err != nil {
		log.Fatalf("Get read mysql error: %s", err)
	}
	writeMysqlPool, writeMysqlResource, writeMysqlDB, err := MysqlDocker()
	if err != nil {
		log.Fatalf("Get write mysql error: %s", err)
	}
	mockReadWriteMysql(cast.ToInt(readMysqlResource.GetPort("3306/tcp")), cast.ToInt(writeMysqlResource.GetPort("3306/tcp")))
	mysqlDB, err := mysqlDockerDB(writeMysqlPool, false)
	if err != nil {
		log.Fatalf("Get mysql gorm error: %s", err)
	}

	readPostgresqlPool, readPostgresqlResource, readPostgresqlDB, err := PostgresqlDocker()
	if err != nil {
		log.Fatalf("Get read postgresql error: %s", err)
	}
	writePostgresqlPool, writePostgresqlResource, writePostgresqlDB, err := PostgresqlDocker()
	if err != nil {
		log.Fatalf("Get write postgresql error: %s", err)
	}
	mockReadWritePostgresql(cast.ToInt(readPostgresqlResource.GetPort("5432/tcp")), cast.ToInt(writePostgresqlResource.GetPort("5432/tcp")))
	postgresqlDB, err := postgresqlDockerDB(writePostgresqlPool, false)
	if err != nil {
		log.Fatalf("Get postgresql gorm error: %s", err)
	}

	_, _, readSqliteDB, err := SqliteDocker(dbDatabase)
	if err != nil {
		log.Fatalf("Get read sqlite error: %s", err)
	}
	writeSqlitePool, _, writeSqliteDB, err := SqliteDocker(dbDatabase1)
	if err != nil {
		log.Fatalf("Get write sqlite error: %s", err)
	}
	mockReadWriteSqlite()
	sqliteDB, err := sqliteDockerDB(writeSqlitePool, false)
	if err != nil {
		log.Fatalf("Get sqlite gorm error: %s", err)
	}

	readSqlserverPool, readSqlserverResource, readSqlserverDB, err := SqlserverDocker()
	if err != nil {
		log.Fatalf("Get read sqlserver error: %s", err)
	}
	writeSqlserverPool, writeSqlserverResource, writeSqlserverDB, err := SqlserverDocker()
	if err != nil {
		log.Fatalf("Get write sqlserver error: %s", err)
	}
	mockReadWriteSqlserver(cast.ToInt(readSqlserverResource.GetPort("1433/tcp")), cast.ToInt(writeSqlserverResource.GetPort("1433/tcp")))
	sqlserverDB, err := sqlserverDockerDB(writeSqlserverPool, false)
	if err != nil {
		log.Fatalf("Get sqlserver gorm error: %s", err)
	}

	dbs := map[contractsorm.Driver]map[string]contractsorm.Query{
		contractsorm.DriverMysql: {
			"mix":   mysqlDB,
			"read":  readMysqlDB,
			"write": writeMysqlDB,
		},
		contractsorm.DriverPostgresql: {
			"mix":   postgresqlDB,
			"read":  readPostgresqlDB,
			"write": writePostgresqlDB,
		},
		contractsorm.DriverSqlite: {
			"mix":   sqliteDB,
			"read":  readSqliteDB,
			"write": writeSqliteDB,
		},
		contractsorm.DriverSqlserver: {
			"mix":   sqlserverDB,
			"read":  readSqlserverDB,
			"write": writeSqlserverDB,
		},
	}

	for drive, db := range dbs {
		t.Run(drive.String(), func(t *testing.T) {
			user := User{Name: "user"}
			assert.Nil(t, db["mix"].Create(&user))
			assert.True(t, user.ID > 0)

			var user2 User
			assert.Nil(t, db["mix"].Find(&user2, user.ID))
			assert.True(t, user2.ID == 0)

			var user3 User
			assert.Nil(t, db["read"].Find(&user3, user.ID))
			assert.True(t, user3.ID == 0)

			var user4 User
			assert.Nil(t, db["write"].Find(&user4, user.ID))
			assert.True(t, user4.ID > 0)
		})
	}

	file.Remove(dbDatabase)
	file.Remove(dbDatabase1)

	if err := readMysqlPool.Purge(readMysqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := writeMysqlPool.Purge(writeMysqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := readPostgresqlPool.Purge(readPostgresqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := writePostgresqlPool.Purge(writePostgresqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := readSqlserverPool.Purge(readSqlserverResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := writeSqlserverPool.Purge(writeSqlserverResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func TestTablePrefixAndSingular(t *testing.T) {
	mysqlPool, mysqlResource, err := initMysqlDocker()
	if err != nil {
		log.Fatalf("Init mysql docker error: %s", err)
	}
	mockMysqlWithPrefixAndSingular(cast.ToInt(mysqlResource.GetPort("3306/tcp")))
	mysqlDB, err := mysqlDockerDBWithPrefixAndSingular(mysqlPool)
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}

	postgresqlPool, postgresqlResource, err := initPostgresqlDocker()
	if err != nil {
		log.Fatalf("Init postgresql docker error: %s", err)
	}
	mockPostgresqlWithPrefixAndSingular(cast.ToInt(postgresqlResource.GetPort("5432/tcp")))
	postgresqlDB, err := postgresqlDockerDBWithPrefixAndSingular(postgresqlPool)
	if err != nil {
		log.Fatalf("Init postgresql error: %s", err)
	}

	sqlitePool, _, err := initSqliteDocker()
	if err != nil {
		log.Fatalf("Init sqlite docker error: %s", err)
	}
	mockSqliteWithPrefixAndSingular(dbDatabase)
	sqliteDB, err := sqliteDockerDBWithPrefixAndSingular(sqlitePool)
	if err != nil {
		log.Fatalf("Init sqlite error: %s", err)
	}

	sqlserverPool, sqlserverResource, err := initSqlserverDocker()
	if err != nil {
		log.Fatalf("Init sqlserver docker error: %s", err)
	}
	mockSqlserverWithPrefixAndSingular(cast.ToInt(sqlserverResource.GetPort("1433/tcp")))
	sqlserverDB, err := sqlserverDockerDBWithPrefixAndSingular(sqlserverPool)
	if err != nil {
		log.Fatalf("Init sqlserver error: %s", err)
	}

	dbs := map[contractsorm.Driver]contractsorm.Query{
		contractsorm.DriverMysql:      mysqlDB,
		contractsorm.DriverPostgresql: postgresqlDB,
		contractsorm.DriverSqlite:     sqliteDB,
		contractsorm.DriverSqlserver:  sqlserverDB,
	}

	for drive, db := range dbs {
		t.Run(drive.String(), func(t *testing.T) {
			user := User{Name: "user"}
			assert.Nil(t, db.Create(&user))
			assert.True(t, user.ID > 0)

			var user1 User
			assert.Nil(t, db.Find(&user1, user.ID))
			assert.True(t, user1.ID > 0)
		})
	}

	file.Remove(dbDatabase)

	if err := mysqlPool.Purge(mysqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := postgresqlPool.Purge(postgresqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := sqlserverPool.Purge(sqlserverResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func paginator(page string, limit string) func(methods contractsorm.Query) contractsorm.Query {
	return func(query contractsorm.Query) contractsorm.Query {
		page, _ := strconv.Atoi(page)
		limit, _ := strconv.Atoi(limit)
		offset := (page - 1) * limit

		return query.Offset(offset).Limit(limit)
	}
}
