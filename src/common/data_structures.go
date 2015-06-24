package common

import (
	//"labix.org/v2/mgo/bson"
//	mgo "gopkg.in/mgo.v2"
	bson "gopkg.in/mgo.v2/bson"
	
	"time"
)

// common data structures that might be shared between front end web servers and backend workers


type SignupUser struct {
	Id              	bson.ObjectId   `bson:"_id" json:"id"`
	CreateTime time.Time     					`bson:"create_time" json:"-"`
	Email string        							`bson:"email" json:"email"`
	AccountState 	string 							`bson:"account_state" json:"-"`   // "unconfirmed","confirmed"
	HashedPassword  string            `bson:"hashed_password" json:"-"`
	Quota 		int64 									`bson:"quota" json:"quota"`// in bytes
	UsedSpace int64										`bson:"used_space" json:"-"` // number of bytes used up out of total quota
	ProjectList []bson.ObjectId				`bson:"project_list" json:"-"`

}

type ContactForm struct {
	Id              	bson.ObjectId   `bson:"_id" json:"id"`
	Email string        							`bson:"email" json:"email"`
	Feedback string      							`bson:"feedback" json:"feedback"`
	Name string      									`bson:"name" json:"name"`
	Subject string      							`bson:"subject" json:"subject"`
	CreateTime time.Time     					`bson:"create_time" json;"-"`

}



type Project struct {
	Id 	bson.ObjectId   							`bson:"_id" json:"id"`
	CreateTime time.Time     					`bson:"create_time" json:"-"`
	Name string 			   							`bson:"name" json:"name"`
	BallName string 									`bson:"ball_name" json:"-"` // tarball name as stored in DIR_PROJECTS_UPLOADED
	FetchUrl string				 						`bson:"fetch_url" json:"fetch_url"` // such as github link or tarball download link
	WorkDir string 										`bson:"work_dir" json:"-"` // directory of source code (i.e. content of BallName is extracted here)
	Status string 										`bson:"status" json:"status"` 		// "uploaded", "extracting"
																																			// "url_received", "fetching"
																																			// "tree_scanning"
																																			// "analyzing"
																																			// "ready"
																																			// "error"

	TotalSize int64 									`bson:"total_size" json:"total_size"` // total size on disk - updated during tree_scanning
	Message	string										`bson:"message" json:"message"`
	//  e.g. when backend processing fails, it should update this message so user knows what error did happen
	
	UserId bson.ObjectId							`bson:"user_id" json:"user_id"` //
	ViewPermission string 						`bson:"view_permission" json:"view_permission"` // "private", "public"
	
}

// representing materialized path: http://docs.mongodb.org/manual/tutorial/model-tree-structures-with-materialized-paths/
type TreePath struct {
	Id 	bson.ObjectId   							`bson:"_id" json:"-"`
	Name 	string								`bson:"name" json:"name"`
	ProjectId bson.ObjectId 		`bson:"project_id" json:"-"`
	Path string												`bson:"path" json:"path"`
	FullPath string							`bson:"full_path" json:"full_path"` // equal to Path+"/"+Name
	IsDir	bool									`bson:"is_dir" json:"is_dir"`
	Size int64									`bson:"size" json:"size"`
}


type Tag struct {
	Id 	bson.ObjectId   							`bson:"_id" json:"id"`
	ProjectId 	bson.ObjectId   							`bson:"project_id" json:"-"`
	Tag 	string   							`bson:"tag" json:"tag"`
}

type TagLocation struct {
	Id 	bson.ObjectId   							`bson:"_id" json:"-"`
	TagId bson.ObjectId									`bson:"tag_id" json:"-"`
	Path string 					`bson:"path" json:"path"`
	LineImage		string		`bson:"line_image" json:"line_image"`
	LineNumber int64 			`bson:"line_number" json:"line_number"`
}


// only check if the format is valid, not checking against database
func IsValidObjectIdHex (objectIdHex string) bool{
	if len (objectIdHex) !=24 {
		return false

	}
	return true

}
