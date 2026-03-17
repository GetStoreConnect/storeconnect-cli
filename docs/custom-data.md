# Custom Data and Custom Objects

StoreConnect supports **custom data fields** on standard objects and fully **custom objects** defined in Salesforce. These enable stores to extend their data model beyond the built-in fields and access that data in Liquid templates.

## Overview

| Feature | Description | Example |
|---------|-------------|---------|
| **Custom Data** | Additional fields on standard objects (Products, Contacts, Accounts, etc.) | A "Fabric Type" field on Products |
| **Custom Objects** | Entirely new object types defined in Salesforce | A "Warranty" or "Book" object |

Both are configured in Salesforce and automatically become available in your Liquid templates for querying and display.

## Custom Data Fields

Custom data fields let you add extra information to standard StoreConnect objects like Products, Contacts, Orders, and more.

### Accessing Custom Data in Templates

Custom data fields are accessed via the `.data` property on any object:

```liquid
{{ product.data.fabric_type }}
{{ product.data.care_instructions }}
{{ contact.data.loyalty_tier }}
{{ account.data.tax_exempt_id }}
```

### Example: Displaying Product Custom Fields

```liquid
{% for product in collection.products %}
  <div class="product-card">
    <h2>{{ product.name }}</h2>
    <p>Price: {{ product.price | money }}</p>

    {% if product.data.fabric_type %}
      <p>Fabric: {{ product.data.fabric_type }}</p>
    {% endif %}

    {% if product.data.care_instructions %}
      <p>Care: {{ product.data.care_instructions }}</p>
    {% endif %}
  </div>
{% endfor %}
```

### Querying by Custom Data

You can filter queries using custom data fields with the `data.` prefix:

```liquid
{% query product2 as silk_products, data.fabric_type: "Silk" %}
  <h2>Silk Products</h2>
  {% for product in silk_products %}
    <p>{{ product.name }}</p>
  {% endfor %}
{% endquery %}
```

### Data Types

Custom data fields support various data types. The value is automatically converted based on how the field is configured in Salesforce:

| Salesforce Type | Liquid Type | Example |
|-----------------|-------------|---------|
| String, Text, Textarea | String | `"Hello World"` |
| Integer, Long | Number | `42` |
| Double, Currency, Percent | Decimal | `19.99` |
| Boolean | Boolean | `true` / `false` |
| Date, DateTime | String (ISO format) | `"2024-01-15"` |
| Picklist | String | `"Option A"` |
| Multipicklist | Array | `["Option A", "Option B"]` |
| Address | Object | See below |
| Location | Object | See below |

#### Address Fields

Address-type custom data fields provide structured access:

```liquid
{{ contact.data.shipping_address.street }}
{{ contact.data.shipping_address.city }}
{{ contact.data.shipping_address.state }}
{{ contact.data.shipping_address.postal_code }}
{{ contact.data.shipping_address.country }}
```

#### Location Fields

Location-type custom data fields provide coordinates:

```liquid
{{ store_location.data.shipping_address.latitude }}
{{ store_location.data.shipping_address.longitude }}
```

## Custom Objects

Custom objects are entirely new data types defined in your Salesforce org. Once configured, they become queryable in Liquid templates just like standard objects.

### Querying Custom Objects

Custom objects are queried using their Salesforce API name (e.g., `book__c`):

```liquid
{% query book__c as books %}
  <h2>Our Book Collection</h2>
  {% for book in books %}
    <div class="book">
      <h3>{{ book.data.title__c }}</h3>
      <p>Author: {{ book.data.author__c }}</p>
      <p>Published: {{ book.data.year__c }}</p>
    </div>
  {% endfor %}
{% endquery %}
```

### Filtering Custom Objects

Apply filters just like standard objects:

```liquid
{% query book__c as fiction_books, data.genre__c: "Fiction" %}
  {% for book in fiction_books %}
    <p>{{ book.data.title__c }}</p>
  {% endfor %}
{% endquery %}
```

### Ordering Results

Use the `order by` clause to sort results:

```liquid
{% query book__c as recent_books order by 'year__c desc' %}
  {% for book in recent_books %}
    <p>{{ book.data.title__c }} ({{ book.data.year__c }})</p>
  {% endfor %}
{% endquery %}
```

### Combining Filters and Ordering

```liquid
{% query book__c as books, data.in_stock__c: true order by 'title__c asc' %}
  {% for book in books %}
    <p>{{ book.data.title__c }}</p>
  {% endfor %}
{% endquery %}
```

## Query Operators

When filtering, you can use comparison operators:

| Operator | Description | Example |
|----------|-------------|---------|
| (none) | Equals | `data.price: 100` |
| `>` | Greater than | `data.price: "> 50"` |
| `>=` | Greater than or equal | `data.price: ">= 50"` |
| `<` | Less than | `data.price: "< 100"` |
| `<=` | Less than or equal | `data.price: "<= 100"` |
| `%...%` | Contains | `data.name: "%shirt%"` |
| `%...` | Ends with | `data.sku: "%XL"` |
| `...%` | Starts with | `data.sku: "PRD%"` |

### Examples with Operators

```liquid
{% query book__c as expensive_books, data.price__c: ">= 50" %}
  <h3>Premium Books ($50+)</h3>
  {% for book in expensive_books %}
    <p>{{ book.data.title__c }} - ${{ book.data.price__c }}</p>
  {% endfor %}
{% endquery %}

{% query product2 as matching_products, data.description: "%organic%" %}
  <h3>Organic Products</h3>
  {% for product in matching_products %}
    <p>{{ product.name }}</p>
  {% endfor %}
{% endquery %}
```

## Salesforce Setup

Custom data and custom objects must be configured in Salesforce before they appear in your templates. This is typically done by your Salesforce administrator.

### Setting Up Custom Data Fields

1. **Create the Custom Field in Salesforce**
   - Navigate to the object (e.g., Product2, Contact)
   - Add a new custom field with your desired data type
   - Note the API name (e.g., `Fabric_Type__c`)

2. **Create a Custom Data Mapping Record**
   - Object: `CustomDataMapping__c`
   - Required fields:
     - **Object API Name**: The Salesforce object (e.g., `product2`, `contact`)
     - **Field API Name**: Your custom field's API name (e.g., `fabric_type__c`)
     - **Data Type**: The field type (string, integer, boolean, etc.)
     - **Access Level**: `read` or `read_write`

3. **Sync Data via Heroku Connect**
   - Ensure the `custom_data` field mapping is enabled
   - Custom field values sync automatically to the website

### Setting Up Custom Objects

1. **Create the Custom Object in Salesforce**
   - Create a new custom object (e.g., `Book__c`)
   - StoreConnect prefixes custom objects with `` (e.g., `book__c`)
   - Add your custom fields to the object

2. **Create Custom Data Mapping Records**
   - Create a mapping for each field you want accessible:
     - **Object API Name**: `book__c`
     - **Field API Name**: `title__c`
     - **Data Type**: `string`
     - **Access Level**: `read`

3. **Enable Sync**
   - Configure Heroku Connect to sync your custom object
   - Records will appear in the `custom_objects` table

### Access Levels

| Level | Description |
|-------|-------------|
| `read` | Field is read-only in templates |
| `read_write` | Field can be updated via forms |

### Indexing Custom Fields

For frequently-queried custom data fields, your administrator can enable indexing:

1. Set the **Indexed** checkbox on the CustomDataMapping record
2. A database index is created automatically
3. Improves query performance for filtered searches

## Common Patterns

### Conditional Display Based on Custom Data

```liquid
{% if product.data.is_featured %}
  <span class="badge">Featured</span>
{% endif %}

{% if product.data.discount_percent > 0 %}
  <span class="sale">{{ product.data.discount_percent }}% off!</span>
{% endif %}
```

### Grouping Products by Custom Field

```liquid
{% query product2 as tshirts, data.category: "T-Shirts" %}
{% query product2 as hoodies, data.category: "Hoodies" %}

<section>
  <h2>T-Shirts</h2>
  {% for product in tshirts %}
    {% include 'snippets/product-card' %}
  {% endfor %}
</section>

<section>
  <h2>Hoodies</h2>
  {% for product in hoodies %}
    {% include 'snippets/product-card' %}
  {% endfor %}
</section>
```

### Building Navigation from Custom Objects

```liquid
{% query brand__c as brands order by 'name__c asc' %}
<nav class="brand-filter">
  <h3>Shop by Brand</h3>
  <ul>
    {% for brand in brands %}
      <li>
        <a href="/collections/all?brand={{ brand.data.slug__c }}">
          {{ brand.data.name__c }}
        </a>
      </li>
    {% endfor %}
  </ul>
</nav>
```

### Displaying Related Custom Objects

```liquid
{% query warranty__c as warranties, data.product_id__c: product.sfid %}
{% if warranties.size > 0 %}
  <div class="warranty-info">
    <h3>Available Warranties</h3>
    {% for warranty in warranties %}
      <div class="warranty-option">
        <strong>{{ warranty.data.name__c }}</strong>
        <p>{{ warranty.data.description__c }}</p>
        <p>Duration: {{ warranty.data.duration_months__c }} months</p>
        <p>Price: {{ warranty.data.price__c | money }}</p>
      </div>
    {% endfor %}
  </div>
{% endif %}
```

## Troubleshooting

### Field Not Appearing

If a custom data field isn't showing in your templates:

1. **Check the CustomDataMapping exists** - The field must be mapped in Salesforce
2. **Verify the field API name** - Names are case-insensitive but must match exactly
3. **Confirm data is synced** - Check that Heroku Connect is running and the field has values
4. **Check access level** - Ensure the field has at least `read` access

### Query Returns No Results

1. **Verify the object API name** - Custom objects use the `` prefix
2. **Check filter values** - Ensure the filter values match actual data
3. **Confirm records exist** - Verify data exists in Salesforce and has synced

### Performance Issues

For slow queries:

1. **Enable indexing** - Set the Indexed flag on frequently-filtered CustomDataMapping records
2. **Limit results** - Use filters to reduce the result set
3. **Cache where possible** - Store query results in variables for reuse
