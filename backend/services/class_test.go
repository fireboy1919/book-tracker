package services

import (
	"testing"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ClassServiceTestSuite struct {
	suite.Suite
	classService *ClassService
	testUser     *models.User
	testTeacher  *models.User
	testAdmin    *models.User
}

func (suite *ClassServiceTestSuite) SetupTest() {
	// Setup test database before each test
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB
	suite.classService = NewClassService(config.TestDB)

	// Create test users
	suite.testUser = &models.User{
		Email:         "user@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "User",
		IsAdmin:       false,
		IsTeacher:     false,
		EmailVerified: true,
	}
	config.DB.Create(suite.testUser)

	suite.testTeacher = &models.User{
		Email:         "teacher@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "Teacher",
		IsAdmin:       false,
		IsTeacher:     true,
		EmailVerified: true,
	}
	config.DB.Create(suite.testTeacher)

	suite.testAdmin = &models.User{
		Email:         "admin@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "Admin",
		IsAdmin:       true,
		IsTeacher:     false,
		EmailVerified: true,
	}
	config.DB.Create(suite.testAdmin)
}

func (suite *ClassServiceTestSuite) TearDownTest() {
	// Cleanup test database after each test
	config.CleanupTestDatabase()
}

func (suite *ClassServiceTestSuite) TestCreateClass_Success_Teacher() {
	req := models.CreateClassRequest{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
	}

	class, err := suite.classService.CreateClass(suite.testTeacher.ID, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), class)
	assert.Equal(suite.T(), "Test Class", class.Name)
	assert.Equal(suite.T(), "A test class", class.Description)
	assert.Equal(suite.T(), 10, class.StudentBooksGoal)
	assert.Equal(suite.T(), 5, class.OtherBooksGoal)
	assert.Equal(suite.T(), suite.testTeacher.ID, class.CreatedByID)

	// Verify teacher was automatically added as member
	var membership models.ClassMembership
	err = config.DB.Where("class_id = ? AND user_id = ? AND role = ?", class.ID, suite.testTeacher.ID, "TEACHER").First(&membership).Error
	assert.NoError(suite.T(), err)
}

func (suite *ClassServiceTestSuite) TestCreateClass_Success_Admin() {
	req := models.CreateClassRequest{
		Name:             "Admin Class",
		Description:      "An admin class",
		StudentBooksGoal: 15,
		OtherBooksGoal:   8,
	}

	class, err := suite.classService.CreateClass(suite.testAdmin.ID, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), class)
	assert.Equal(suite.T(), "Admin Class", class.Name)
}

func (suite *ClassServiceTestSuite) TestCreateClass_Failure_NotTeacherOrAdmin() {
	req := models.CreateClassRequest{
		Name:             "User Class",
		Description:      "A user class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
	}

	class, err := suite.classService.CreateClass(suite.testUser.ID, req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), class)
	assert.Contains(suite.T(), err.Error(), "only teachers and admins can create classes")
}

func (suite *ClassServiceTestSuite) TestGetClasses_Admin() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	classes, err := suite.classService.GetClasses(suite.testAdmin.ID, true)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), classes, 1)
	assert.Equal(suite.T(), "Test Class", classes[0].Name)
}

func (suite *ClassServiceTestSuite) TestGetClasses_Teacher() {
	// Create a test class and add teacher as member
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testAdmin.ID,
	}
	config.DB.Create(&testClass)

	membership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&membership)

	classes, err := suite.classService.GetClasses(suite.testTeacher.ID, false)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), classes, 1)
	assert.Equal(suite.T(), "Test Class", classes[0].Name)
}

func (suite *ClassServiceTestSuite) TestGetClasses_NoMembership() {
	// Create a test class without adding user as member
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testAdmin.ID,
	}
	config.DB.Create(&testClass)

	classes, err := suite.classService.GetClasses(suite.testUser.ID, false)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), classes, 0)
}

func (suite *ClassServiceTestSuite) TestAddClassMember_Success() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as class member first
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	req := models.AddClassMemberRequest{
		UserID: suite.testUser.ID,
		Role:   "STUDENT",
	}

	membership, err := suite.classService.AddClassMember(testClass.ID, suite.testTeacher.ID, false, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), membership)
	assert.Equal(suite.T(), testClass.ID, membership.ClassID)
	assert.Equal(suite.T(), suite.testUser.ID, membership.UserID)
	assert.Equal(suite.T(), "STUDENT", membership.Role)
}

func (suite *ClassServiceTestSuite) TestAddClassMember_TeacherRole_RequiresTeacherUser() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as class member first
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	req := models.AddClassMemberRequest{
		UserID: suite.testUser.ID, // Regular user, not teacher
		Role:   "TEACHER",
	}

	membership, err := suite.classService.AddClassMember(testClass.ID, suite.testTeacher.ID, false, req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), membership)
	assert.Contains(suite.T(), err.Error(), "user must be a teacher to be assigned TEACHER role")
}

func (suite *ClassServiceTestSuite) TestAddClassMember_DuplicateMembership() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as class member first
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	// Add user as student
	existingMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testUser.ID,
		Role:    "STUDENT",
	}
	config.DB.Create(&existingMembership)

	req := models.AddClassMemberRequest{
		UserID: suite.testUser.ID,
		Role:   "STUDENT",
	}

	membership, err := suite.classService.AddClassMember(testClass.ID, suite.testTeacher.ID, false, req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), membership)
	assert.Contains(suite.T(), err.Error(), "user is already a member of this class")
}

func (suite *ClassServiceTestSuite) TestGetClassStudents_Success() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as class member
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	// Add students
	student1 := models.User{
		Email:         "student1@example.com",
		FirstName:     "Alice",
		LastName:      "Smith",
		EmailVerified: true,
	}
	config.DB.Create(&student1)

	student2 := models.User{
		Email:         "student2@example.com",
		FirstName:     "Bob",
		LastName:      "Johnson",
		EmailVerified: true,
	}
	config.DB.Create(&student2)

	studentMembership1 := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  student1.ID,
		Role:    "STUDENT",
	}
	config.DB.Create(&studentMembership1)

	studentMembership2 := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  student2.ID,
		Role:    "STUDENT",
	}
	config.DB.Create(&studentMembership2)

	students, err := suite.classService.GetClassStudents(testClass.ID, suite.testTeacher.ID, false)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), students, 2)
	// Should be sorted alphabetically by last name
	assert.Equal(suite.T(), "Bob", students[0].FirstName)
	assert.Equal(suite.T(), "Alice", students[1].FirstName)
}

func (suite *ClassServiceTestSuite) TestGetClassStudents_AccessDenied() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	students, err := suite.classService.GetClassStudents(testClass.ID, suite.testUser.ID, false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), students)
	assert.Contains(suite.T(), err.Error(), "access denied to this class")
}

func (suite *ClassServiceTestSuite) TestAssignChildToClass_Success() {
	// Create test child
	testChild := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
		OwnerID:   suite.testUser.ID,
	}
	config.DB.Create(&testChild)

	// Create test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	child, err := suite.classService.AssignChildToClass(testChild.ID, testClass.ID, suite.testUser.ID, false)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), child)
	assert.Equal(suite.T(), testChild.ID, child.ID)

	// Verify child was assigned to class
	var updatedChild models.Child
	config.DB.First(&updatedChild, testChild.ID)
	assert.NotNil(suite.T(), updatedChild.ClassID)
	assert.Equal(suite.T(), testClass.ID, *updatedChild.ClassID)
}

func (suite *ClassServiceTestSuite) TestAssignChildToClass_NotOwner() {
	// Create test child owned by different user
	testChild := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
		OwnerID:   suite.testTeacher.ID, // Different owner
	}
	config.DB.Create(&testChild)

	// Create test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	child, err := suite.classService.AssignChildToClass(testChild.ID, testClass.ID, suite.testUser.ID, false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), child)
	assert.Contains(suite.T(), err.Error(), "you don't have permission to assign this child to a class")
}

func (suite *ClassServiceTestSuite) TestUpdateClass_Success() {
	// Create a test class
	testClass := models.Class{
		Name:             "Old Name",
		Description:      "Old description",
		StudentBooksGoal: 5,
		OtherBooksGoal:   3,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as class member
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	req := models.UpdateClassRequest{
		Name:             "New Name",
		Description:      "New description",
		StudentBooksGoal: 10,
		OtherBooksGoal:   8,
	}

	class, err := suite.classService.UpdateClass(testClass.ID, suite.testTeacher.ID, false, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), class)
	assert.Equal(suite.T(), "New Name", class.Name)
	assert.Equal(suite.T(), "New description", class.Description)
	assert.Equal(suite.T(), 10, class.StudentBooksGoal)
	assert.Equal(suite.T(), 8, class.OtherBooksGoal)
}

func (suite *ClassServiceTestSuite) TestRemoveClassMember_Success() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as class member
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	// Add user as student
	studentMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testUser.ID,
		Role:    "STUDENT",
	}
	config.DB.Create(&studentMembership)

	err := suite.classService.RemoveClassMember(testClass.ID, suite.testUser.ID, suite.testTeacher.ID, false)

	assert.NoError(suite.T(), err)

	// Verify membership was removed
	var membership models.ClassMembership
	err = config.DB.Where("class_id = ? AND user_id = ?", testClass.ID, suite.testUser.ID).First(&membership).Error
	assert.Error(suite.T(), err) // Should not find the record
}

func (suite *ClassServiceTestSuite) TestGetAvailableClasses() {
	// Create test classes
	testClass1 := models.Class{
		Name:             "Class 1",
		Description:      "First class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass1)

	testClass2 := models.Class{
		Name:             "Class 2",
		Description:      "Second class",
		StudentBooksGoal: 15,
		OtherBooksGoal:   7,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass2)

	classes, err := suite.classService.GetAvailableClasses()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), classes, 2)
	assert.Equal(suite.T(), "Class 1", classes[0].Name)
	assert.Equal(suite.T(), "Class 2", classes[1].Name)
}

func (suite *ClassServiceTestSuite) TestGetClassTeachers_Success() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as class member
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	// Create another teacher
	anotherTeacher := &models.User{
		Email:         "teacher2@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Another",
		LastName:      "Teacher",
		IsAdmin:       false,
		IsTeacher:     true,
		EmailVerified: true,
	}
	config.DB.Create(anotherTeacher)

	// Add another teacher as class member
	anotherTeacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  anotherTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&anotherTeacherMembership)

	teachers, err := suite.classService.GetClassTeachers(testClass.ID, suite.testTeacher.ID, false)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), teachers, 2)
	// Should be sorted alphabetically by last name
	assert.Equal(suite.T(), "Teacher", teachers[0].LastName) // "Another Teacher"
	assert.Equal(suite.T(), "Teacher", teachers[1].LastName) // "Test Teacher"
}

func (suite *ClassServiceTestSuite) TestAssignChildToClass_RestrictsParentChangeOnceAssigned() {
	// Create a test child owned by regular user
	testChild := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "1st",
		OwnerID:   suite.testUser.ID,
	}
	config.DB.Create(&testChild)

	// Create test classes
	testClass1 := models.Class{
		Name:             "Test Class 1",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass1)

	testClass2 := models.Class{
		Name:             "Test Class 2",
		Description:      "Another test class",
		StudentBooksGoal: 8,
		OtherBooksGoal:   4,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass2)

	// First assignment should work (parent can assign initially)
	child, err := suite.classService.AssignChildToClass(testChild.ID, testClass1.ID, suite.testUser.ID, false)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), child)
	assert.Equal(suite.T(), &testClass1.ID, child.ClassID)

	// Second assignment by same parent should fail (only teachers can change)
	child2, err := suite.classService.AssignChildToClass(testChild.ID, testClass2.ID, suite.testUser.ID, false)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), child2)
	assert.Contains(suite.T(), err.Error(), "only teachers can change class assignments once a child is already assigned to a class")

	// But teacher should be able to change it
	child3, err := suite.classService.AssignChildToClass(testChild.ID, testClass2.ID, suite.testTeacher.ID, false)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), child3)
	assert.Equal(suite.T(), &testClass2.ID, child3.ClassID)
}

func (suite *ClassServiceTestSuite) TestAssignChildToClass_AdminCanAlwaysChange() {
	// Create a test child owned by regular user
	testChild := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "1st",
		OwnerID:   suite.testUser.ID,
	}
	config.DB.Create(&testChild)

	// Create test classes
	testClass1 := models.Class{
		Name:             "Test Class 1",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass1)

	testClass2 := models.Class{
		Name:             "Test Class 2",
		Description:      "Another test class",
		StudentBooksGoal: 8,
		OtherBooksGoal:   4,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass2)

	// Assign to first class
	suite.classService.AssignChildToClass(testChild.ID, testClass1.ID, suite.testUser.ID, false)

	// Admin should be able to change it even if already assigned
	child, err := suite.classService.AssignChildToClass(testChild.ID, testClass2.ID, suite.testAdmin.ID, true)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), child)
	assert.Equal(suite.T(), &testClass2.ID, child.ClassID)
}

func TestClassServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ClassServiceTestSuite))
}