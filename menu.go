package gwf

type Item struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	IsActive bool   `json:"is_active"`
}

type Menu struct {
	Name     string  `json:"name"`
	IsActive bool    `json:"is_active"`
	ItemList []*Item `json:"item_list"`
}

type MenuList []*Menu

func (ml MenuList) SetActive(url string) {
	for _, menu := range ml {
		for _, item := range menu.ItemList {
			item.IsActive = false
			menu.IsActive = false
		}
	}

	for _, menu := range ml {
		for _, item := range menu.ItemList {
			if item.Url == url {
				item.IsActive = true
				menu.IsActive = true
			}
		}
	}
}

func (ml MenuList) SetUrlPrefix(prefix string) {
	for _, menu := range ml {
		for _, item := range menu.ItemList {
			item.Url = prefix + item.Url
		}
	}
}
