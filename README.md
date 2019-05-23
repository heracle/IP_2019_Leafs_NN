# IP_2019_Leafs_NN

## authenticate  [/authenticate]
JSON device -> server

    type User struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

JSON server -> device

    type JwtToken struct {
        Token string `json:"token"`
    }

## Print info about all classes available  [/all_classes]

none JSON device -> server

JSON server -> device

    type AllClassesInfo struct {
        Data [noClasses]QueryInfo
    }

    type QueryInfo struct {
        ID            string
        Common_name   string
        Specific_name string
        Details       struct {
            Wikipedia string
        }
    }

## Print info about user history  [/get_history]

JSON device -> server

    type JwtToken struct {
        Token string `json:"token"`
    }

JSON server -> device

    type QueryHistory struct {
        Photo  string
        Answer QueryInfo
    }

    type HistoryData struct {
        Data []QueryHistory
    }

    type QueryInfo struct {
        ID            string
        Common_name   string
        Specific_name string
        Details       struct {
            Wikipedia string
        }
    }

## Query to server with an image  [/post]

JSON device -> server

    type jsonClass struct {
        Token string `json:"token"`
        Photo string `json:"photo"`
    }

JSON server -> device

    type QueryInfo struct {
        ID            string
        Common_name   string
        Specific_name string
        Details       struct {
            Wikipedia string
        }
    }