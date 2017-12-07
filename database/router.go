package database

// Router describe basic info about router
type Router struct {
	Name           string `gorm:"column:name"`
	PrivateAddress string `gorm:"column:priv"`
	PublicAddress  string `gorm:"column:public"`
}

// Routers list of Router type
type Routers []Router

// GetRoutersList get from database list of a router to configure.
// All router must be Konfiguracja/Configuration type
func GetRoutersList() (*Routers, error) {
	query := `SELECT n.name, 
	    			 INET_NTOA(n.ipaddr) priv, 
	    			 INET_NTOA(n.ipaddr_pub) public
		        FROM lms.netdevices nd 
		   LEFT JOIN nastypes t on (t.id=nd.nastype) 
		   LEFT JOIN vnodes n ON (n.netdev=nd.netnodeid) 
		       WHERE t.name='Konfiguracja'
			 	 AND n.ownerid=0`

	var ret = new(Routers)
	err := Db.Raw(query).Find(ret).Error
	return ret, err
}
