package services

import (
	"errors"

	"gorm.io/gorm"

	"github.com/booktracker/backend/models"
)

type ClassService struct {
	DB *gorm.DB
}

func NewClassService(db *gorm.DB) *ClassService {
	return &ClassService{DB: db}
}

// CreateClass creates a new class
func (s *ClassService) CreateClass(creatorID uint, req models.CreateClassRequest) (*models.ClassResponse, error) {
	// Check if user is teacher or admin
	var user models.User
	if err := s.DB.First(&user, creatorID).Error; err != nil {
		return nil, err
	}

	if !user.IsTeacher && !user.IsAdmin {
		return nil, errors.New("only teachers and admins can create classes")
	}

	class := models.Class{
		Name:             req.Name,
		Description:      req.Description,
		StudentBooksGoal: req.StudentBooksGoal,
		OtherBooksGoal:   req.OtherBooksGoal,
		CreatedByID:      creatorID,
	}

	if err := s.DB.Create(&class).Error; err != nil {
		return nil, err
	}

	// Automatically add creator as teacher
	membership := models.ClassMembership{
		ClassID: class.ID,
		UserID:  creatorID,
		Role:    "TEACHER",
	}

	if err := s.DB.Create(&membership).Error; err != nil {
		return nil, err
	}

	return &models.ClassResponse{
		ID:               class.ID,
		Name:             class.Name,
		Description:      class.Description,
		StudentBooksGoal: class.StudentBooksGoal,
		OtherBooksGoal:   class.OtherBooksGoal,
		CreatedByID:      class.CreatedByID,
		CreatedAt:        class.CreatedAt,
		UpdatedAt:        class.UpdatedAt,
	}, nil
}

// GetClasses returns classes based on user role
func (s *ClassService) GetClasses(userID uint, isAdmin bool) ([]models.ClassResponse, error) {
	var classes []models.Class

	if isAdmin {
		// Admins can see all classes
		if err := s.DB.Find(&classes).Error; err != nil {
			return nil, err
		}
	} else {
		// Regular users can only see classes they're members of
		if err := s.DB.Joins("JOIN class_memberships ON classes.id = class_memberships.class_id").
			Where("class_memberships.user_id = ?", userID).
			Find(&classes).Error; err != nil {
			return nil, err
		}
	}

	var response []models.ClassResponse
	for _, class := range classes {
		response = append(response, models.ClassResponse{
			ID:               class.ID,
			Name:             class.Name,
			Description:      class.Description,
			StudentBooksGoal: class.StudentBooksGoal,
			OtherBooksGoal:   class.OtherBooksGoal,
			CreatedByID:      class.CreatedByID,
			CreatedAt:        class.CreatedAt,
			UpdatedAt:        class.UpdatedAt,
		})
	}

	return response, nil
}

// GetClass returns a specific class with its members and children
func (s *ClassService) GetClass(classID, userID uint, isAdmin bool) (*models.ClassWithMembersResponse, error) {
	var class models.Class

	// Check if user has access to this class
	if !isAdmin {
		var membership models.ClassMembership
		if err := s.DB.Where("class_id = ? AND user_id = ?", classID, userID).First(&membership).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.New("access denied to this class")
			}
			return nil, err
		}
	}

	if err := s.DB.First(&class, classID).Error; err != nil {
		return nil, err
	}

	// Get members
	var memberships []models.ClassMembership
	if err := s.DB.Preload("User").Where("class_id = ?", classID).Find(&memberships).Error; err != nil {
		return nil, err
	}

	var members []models.ClassMembershipResponse
	for _, membership := range memberships {
		userResponse := &models.UserResponse{
			ID:            membership.User.ID,
			Email:         membership.User.Email,
			FirstName:     membership.User.FirstName,
			LastName:      membership.User.LastName,
			IsAdmin:       membership.User.IsAdmin,
			IsTeacher:     membership.User.IsTeacher,
			EmailVerified: membership.User.EmailVerified,
			CreatedAt:     membership.User.CreatedAt,
		}

		members = append(members, models.ClassMembershipResponse{
			ID:        membership.ID,
			ClassID:   membership.ClassID,
			UserID:    membership.UserID,
			Role:      membership.Role,
			CreatedAt: membership.CreatedAt,
			User:      userResponse,
		})
	}

	// Get children assigned to this class
	var children []models.Child
	if err := s.DB.Where("class_id = ?", classID).Find(&children).Error; err != nil {
		return nil, err
	}

	var childrenResponse []models.ChildResponse
	for _, child := range children {
		childrenResponse = append(childrenResponse, models.ChildResponse{
			ID:        child.ID,
			FirstName: child.FirstName,
			LastName:  child.LastName,
			Grade:     child.Grade,
			OwnerID:   child.OwnerID,
			ClassID:   child.ClassID,
			CreatedAt: child.CreatedAt,
		})
	}

	return &models.ClassWithMembersResponse{
		ClassResponse: models.ClassResponse{
			ID:               class.ID,
			Name:             class.Name,
			Description:      class.Description,
			StudentBooksGoal: class.StudentBooksGoal,
			OtherBooksGoal:   class.OtherBooksGoal,
			CreatedByID:      class.CreatedByID,
			CreatedAt:        class.CreatedAt,
			UpdatedAt:        class.UpdatedAt,
		},
		Members:  members,
		Children: childrenResponse,
	}, nil
}

// UpdateClass updates a class
func (s *ClassService) UpdateClass(classID, userID uint, isAdmin bool, req models.UpdateClassRequest) (*models.ClassResponse, error) {
	var class models.Class
	if err := s.DB.First(&class, classID).Error; err != nil {
		return nil, err
	}

	// Check permissions
	if !isAdmin {
		// Check if user is a teacher in this class
		var membership models.ClassMembership
		if err := s.DB.Where("class_id = ? AND user_id = ? AND role = ?", classID, userID, "TEACHER").First(&membership).Error; err != nil {
			return nil, errors.New("only teachers and admins can update classes")
		}
	}

	class.Name = req.Name
	class.Description = req.Description
	class.StudentBooksGoal = req.StudentBooksGoal
	class.OtherBooksGoal = req.OtherBooksGoal

	if err := s.DB.Save(&class).Error; err != nil {
		return nil, err
	}

	return &models.ClassResponse{
		ID:               class.ID,
		Name:             class.Name,
		Description:      class.Description,
		StudentBooksGoal: class.StudentBooksGoal,
		OtherBooksGoal:   class.OtherBooksGoal,
		CreatedByID:      class.CreatedByID,
		CreatedAt:        class.CreatedAt,
		UpdatedAt:        class.UpdatedAt,
	}, nil
}

// DeleteClass deletes a class (admin only)
func (s *ClassService) DeleteClass(classID, userID uint, isAdmin bool) error {
	if !isAdmin {
		return errors.New("only admins can delete classes")
	}

	var class models.Class
	if err := s.DB.First(&class, classID).Error; err != nil {
		return err
	}

	// Remove all memberships first
	if err := s.DB.Where("class_id = ?", classID).Delete(&models.ClassMembership{}).Error; err != nil {
		return err
	}

	// Remove class assignments from children
	if err := s.DB.Model(&models.Child{}).Where("class_id = ?", classID).Update("class_id", nil).Error; err != nil {
		return err
	}

	// Delete the class
	return s.DB.Delete(&class).Error
}

// AddClassMember adds a user to a class
func (s *ClassService) AddClassMember(classID, requestingUserID uint, isAdmin bool, req models.AddClassMemberRequest) (*models.ClassMembershipResponse, error) {
	// Check if class exists
	var class models.Class
	if err := s.DB.First(&class, classID).Error; err != nil {
		return nil, err
	}

	// Check if user to be added exists
	var targetUser models.User
	if err := s.DB.First(&targetUser, req.UserID).Error; err != nil {
		return nil, err
	}

	// Check permissions
	if !isAdmin {
		// Check if requesting user is a teacher in this class
		var membership models.ClassMembership
		if err := s.DB.Where("class_id = ? AND user_id = ? AND role = ?", classID, requestingUserID, "TEACHER").First(&membership).Error; err != nil {
			return nil, errors.New("only teachers and admins can add members to classes")
		}
	}

	// Check if user is already a member
	var existingMembership models.ClassMembership
	if err := s.DB.Where("class_id = ? AND user_id = ?", classID, req.UserID).First(&existingMembership).Error; err == nil {
		return nil, errors.New("user is already a member of this class")
	}

	// Validate role requirements
	if req.Role == "TEACHER" && !targetUser.IsTeacher && !targetUser.IsAdmin {
		return nil, errors.New("user must be a teacher to be assigned TEACHER role")
	}

	newMembership := models.ClassMembership{
		ClassID: classID,
		UserID:  req.UserID,
		Role:    req.Role,
	}

	if err := s.DB.Create(&newMembership).Error; err != nil {
		return nil, err
	}

	// Load the user data for response
	if err := s.DB.Preload("User").First(&newMembership, newMembership.ID).Error; err != nil {
		return nil, err
	}

	userResponse := &models.UserResponse{
		ID:            newMembership.User.ID,
		Email:         newMembership.User.Email,
		FirstName:     newMembership.User.FirstName,
		LastName:      newMembership.User.LastName,
		IsAdmin:       newMembership.User.IsAdmin,
		IsTeacher:     newMembership.User.IsTeacher,
		EmailVerified: newMembership.User.EmailVerified,
		CreatedAt:     newMembership.User.CreatedAt,
	}

	return &models.ClassMembershipResponse{
		ID:        newMembership.ID,
		ClassID:   newMembership.ClassID,
		UserID:    newMembership.UserID,
		Role:      newMembership.Role,
		CreatedAt: newMembership.CreatedAt,
		User:      userResponse,
	}, nil
}

// RemoveClassMember removes a user from a class
func (s *ClassService) RemoveClassMember(classID, memberID, requestingUserID uint, isAdmin bool) error {
	// Check permissions
	if !isAdmin {
		// Check if requesting user is a teacher in this class
		var membership models.ClassMembership
		if err := s.DB.Where("class_id = ? AND user_id = ? AND role = ?", classID, requestingUserID, "TEACHER").First(&membership).Error; err != nil {
			return errors.New("only teachers and admins can remove members from classes")
		}
	}

	var membershipToRemove models.ClassMembership
	if err := s.DB.Where("class_id = ? AND user_id = ?", classID, memberID).First(&membershipToRemove).Error; err != nil {
		return err
	}

	return s.DB.Delete(&membershipToRemove).Error
}

// GetClassStudents returns all students (children) in a class with book counts, sorted alphabetically
func (s *ClassService) GetClassStudents(classID, userID uint, isAdmin bool) ([]models.ClassStudentResponse, error) {
	// Check if user has access to this class
	if !isAdmin {
		var membership models.ClassMembership
		if err := s.DB.Where("class_id = ? AND user_id = ?", classID, userID).First(&membership).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.New("access denied to this class")
			}
			return nil, err
		}
	}

	// Get the class to access the goals
	var class models.Class
	if err := s.DB.First(&class, classID).Error; err != nil {
		return nil, err
	}

	// Get all children assigned to this class, sorted alphabetically by last name, then first name
	var children []models.Child
	if err := s.DB.Preload("Owner").
		Where("class_id = ?", classID).
		Order("last_name ASC, first_name ASC").
		Find(&children).Error; err != nil {
		return nil, err
	}

	var response []models.ClassStudentResponse
	for _, child := range children {
		// Count books read by student (ReadByParent = false)
		var studentBooksRead int64
		s.DB.Model(&models.Book{}).Where("child_id = ? AND read_by_parent = ?", child.ID, false).Count(&studentBooksRead)

		// Count books read to student by parent (ReadByParent = true)
		var readToBooksRead int64
		s.DB.Model(&models.Book{}).Where("child_id = ? AND read_by_parent = ?", child.ID, true).Count(&readToBooksRead)

		// Check if total reading goal is met
		// ReadTo goal is a maximum - only books up to that limit count toward achievement
		effectiveReadToBooks := int(readToBooksRead)
		if effectiveReadToBooks > class.OtherBooksGoal {
			effectiveReadToBooks = class.OtherBooksGoal
		}
		totalBooksRead := int(studentBooksRead) + effectiveReadToBooks
		goalsReached := totalBooksRead >= class.StudentBooksGoal

		response = append(response, models.ClassStudentResponse{
			ID:               child.ID,
			FirstName:        child.FirstName,
			LastName:         child.LastName,
			OwnerID:          child.OwnerID,
			ClassID:          child.ClassID,
			CreatedAt:        child.CreatedAt,
			StudentBooksRead: int(studentBooksRead),
			ReadToBooksRead:  int(readToBooksRead),
			StudentGoal:      class.StudentBooksGoal,
			ReadToGoal:       class.OtherBooksGoal,
			GoalsReached:     goalsReached,
		})
	}

	return response, nil
}

// RemoveChildFromClass removes a child from a class by setting their class_id to null
func (s *ClassService) RemoveChildFromClass(childID, classID, userID uint, isAdmin bool) error {
	// Get the child first
	var child models.Child
	if err := s.DB.First(&child, childID).Error; err != nil {
		return err
	}

	// Check if child is actually in this class
	if child.ClassID == nil || *child.ClassID != classID {
		return errors.New("child is not in this class")
	}

	// Check permissions - only admins, teachers, or the child's owner can remove them
	if !isAdmin {
		var user models.User
		if err := s.DB.First(&user, userID).Error; err != nil {
			return err
		}

		// Check if user is the child's owner
		if child.OwnerID != userID {
			// Check if user is a teacher in this class
			var membership models.ClassMembership
			if err := s.DB.Where("class_id = ? AND user_id = ? AND role = ?", classID, userID, "TEACHER").First(&membership).Error; err != nil {
				return errors.New("you don't have permission to remove this child from the class")
			}
		}
	}

	// Remove child from class by setting class_id to null
	child.ClassID = nil
	return s.DB.Save(&child).Error
}

// AssignChildToClass assigns a child to a class
func (s *ClassService) AssignChildToClass(childID, classID, userID uint, isAdmin bool) (*models.ChildResponse, error) {
	// Check if child exists and user has permission
	var child models.Child
	if err := s.DB.First(&child, childID).Error; err != nil {
		return nil, err
	}

	// Check if class exists
	var class models.Class
	if err := s.DB.First(&class, classID).Error; err != nil {
		return nil, err
	}

	// Check if user is a teacher
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Check permissions
	if !isAdmin {
		// If child is already assigned to a class, only teachers can change it
		if child.ClassID != nil && *child.ClassID != 0 && !user.IsTeacher {
			return nil, errors.New("only teachers can change class assignments once a child is already assigned to a class")
		}

		// Check if user has access to this child
		// Teachers can assign any child to their classes, others need to own the child or have permission
		if !user.IsTeacher {
			hasAccess := false
			if child.OwnerID == userID {
				hasAccess = true
			} else {
				var permission models.Permission
				if err := s.DB.Where("user_id = ? AND child_id = ? AND permission_type = ?", userID, childID, "EDIT").First(&permission).Error; err == nil {
					hasAccess = true
				}
			}

			if !hasAccess {
				return nil, errors.New("you don't have permission to assign this child to a class")
			}
		}
	}

	child.ClassID = &classID
	if err := s.DB.Save(&child).Error; err != nil {
		return nil, err
	}

	return &models.ChildResponse{
		ID:        child.ID,
		FirstName: child.FirstName,
		LastName:  child.LastName,
		Grade:     child.Grade,
		OwnerID:   child.OwnerID,
		ClassID:   child.ClassID,
		CreatedAt: child.CreatedAt,
	}, nil
}

// GetAvailableClasses returns all classes available for assignment
func (s *ClassService) GetAvailableClasses() ([]models.ClassResponse, error) {
	var classes []models.Class
	if err := s.DB.Find(&classes).Error; err != nil {
		return nil, err
	}

	var response []models.ClassResponse
	for _, class := range classes {
		response = append(response, models.ClassResponse{
			ID:               class.ID,
			Name:             class.Name,
			Description:      class.Description,
			StudentBooksGoal: class.StudentBooksGoal,
			OtherBooksGoal:   class.OtherBooksGoal,
			CreatedByID:      class.CreatedByID,
			CreatedAt:        class.CreatedAt,
			UpdatedAt:        class.UpdatedAt,
		})
	}

	return response, nil
}

// GetClassTeachers returns all teachers in a class, sorted alphabetically
func (s *ClassService) GetClassTeachers(classID, userID uint, isAdmin bool) ([]models.UserResponse, error) {
	// Check if user has access to this class
	if !isAdmin {
		var membership models.ClassMembership
		if err := s.DB.Where("class_id = ? AND user_id = ?", classID, userID).First(&membership).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.New("access denied to this class")
			}
			return nil, err
		}
	}

	// Get all teachers in the class, sorted alphabetically by last name, then first name
	var teachers []models.User
	if err := s.DB.Joins("JOIN class_memberships ON users.id = class_memberships.user_id").
		Where("class_memberships.class_id = ? AND class_memberships.role = ?", classID, "TEACHER").
		Order("users.last_name ASC, users.first_name ASC").
		Find(&teachers).Error; err != nil {
		return nil, err
	}

	var response []models.UserResponse
	for _, teacher := range teachers {
		response = append(response, models.UserResponse{
			ID:            teacher.ID,
			Email:         teacher.Email,
			FirstName:     teacher.FirstName,
			LastName:      teacher.LastName,
			IsAdmin:       teacher.IsAdmin,
			IsTeacher:     teacher.IsTeacher,
			EmailVerified: teacher.EmailVerified,
			CreatedAt:     teacher.CreatedAt,
		})
	}

	return response, nil
}

// GetTeacherClasses returns all classes where the user is a teacher
func (s *ClassService) GetTeacherClasses(userID uint) ([]models.ClassResponse, error) {
	var classes []models.Class

	if err := s.DB.Joins("JOIN class_memberships ON classes.id = class_memberships.class_id").
		Where("class_memberships.user_id = ? AND class_memberships.role = ?", userID, "TEACHER").
		Find(&classes).Error; err != nil {
		return nil, err
	}

	var response []models.ClassResponse
	for _, class := range classes {
		response = append(response, models.ClassResponse{
			ID:               class.ID,
			Name:             class.Name,
			Description:      class.Description,
			StudentBooksGoal: class.StudentBooksGoal,
			OtherBooksGoal:   class.OtherBooksGoal,
			CreatedByID:      class.CreatedByID,
			CreatedAt:        class.CreatedAt,
			UpdatedAt:        class.UpdatedAt,
		})
	}

	return response, nil
}

// SearchStudents searches for students by name for teachers to add to classes
func (s *ClassService) SearchStudents(query string, userID uint, isAdmin bool) ([]models.ChildResponse, error) {
	// Check if user is a teacher or admin
	if !isAdmin {
		var user models.User
		if err := s.DB.First(&user, userID).Error; err != nil {
			return nil, err
		}
		if !user.IsTeacher {
			return nil, errors.New("only teachers and admins can search students")
		}
	}

	// Search children by first name or last name
	var children []models.Child
	searchPattern := "%" + query + "%"
	
	if err := s.DB.Where("LOWER(first_name) LIKE LOWER(?) OR LOWER(last_name) LIKE LOWER(?)", searchPattern, searchPattern).
		Order("last_name ASC, first_name ASC").
		Limit(20). // Limit results for performance
		Find(&children).Error; err != nil {
		return nil, err
	}

	var response []models.ChildResponse
	for _, child := range children {
		response = append(response, models.ChildResponse{
			ID:        child.ID,
			FirstName: child.FirstName,
			LastName:  child.LastName,
			Grade:     child.Grade,
			OwnerID:   child.OwnerID,
			ClassID:   child.ClassID,
			CreatedAt: child.CreatedAt,
		})
	}

	return response, nil
}