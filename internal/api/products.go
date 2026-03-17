package api

// Product represents a StoreConnect product
type Product struct {
	SCID        string                 `json:"sc_id"`
	SFID        string                 `json:"sfid,omitempty"`
	Name        string                 `json:"name"`
	SKU         string                 `json:"sku,omitempty"`
	Description string                 `json:"description,omitempty"`
	Price       float64                `json:"price,omitempty"`
	CustomData  map[string]interface{} `json:"custom_data,omitempty"`
}

// Products handles product-related endpoints
type Products struct {
	client *Client
}

// NewProducts creates a new Products service
func NewProducts(client *Client) *Products {
	return &Products{client: client}
}

// List returns all products
func (p *Products) List() ([]Product, error) {
	var result struct {
		Products []Product `json:"products"`
	}

	err := p.client.Get("/api/v1/products", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Products, nil
}

// Get retrieves a single product by ID
func (p *Products) Get(id string) (*Product, error) {
	var result Product
	err := p.client.Get("/api/v1/products/"+id, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
