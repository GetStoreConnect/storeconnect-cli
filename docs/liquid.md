# StoreConnect Liquid Reference

This comprehensive guide covers the Liquid templating language as implemented in StoreConnect. Use this reference to build dynamic, data-driven storefronts without writing backend code.

## Table of Contents

1. [Introduction](#introduction)
2. [Liquid Basics](#liquid-basics)
3. [Template Types](#template-types)
4. [Global Objects](#global-objects)
5. [Liquid Tags](#liquid-tags)
6. [Liquid Filters](#liquid-filters)
7. [Liquid Controllers](#liquid-controllers)
8. [Forms](#forms)
9. [Working with Data](#working-with-data)
10. [Best Practices](#best-practices)

---

## Introduction

Liquid is a templating language that lets you output dynamic content on your storefront. It uses a combination of **objects**, **tags**, and **filters** to load and display store data.

### Key Concepts

- **Objects** contain content that Liquid displays on a page (e.g., `{{ product.name }}`)
- **Tags** create logic and control flow (e.g., `{% if %}`, `{% for %}`)
- **Filters** modify output (e.g., `{{ product.price | money }}`)

### Syntax

```liquid
{%- comment -%}
  This is a comment
  Use {%- and -%} to strip whitespace
{%- endcomment -%}

{%- assign variable = "value" -%}
{{ variable }}  {%- comment -%} Outputs: value {%- endcomment -%}

{%- if condition -%}
  True branch
{%- else -%}
  False branch
{%- endif -%}
```

---

## Liquid Basics

### Output

Use double curly braces to output data:

```liquid
{{ product.name }}
{{ product.price | money }}
{{ current_store.name }}
```

### Variables

Assign variables using `{% assign %}`:

```liquid
{%- assign sale_price = product.price | times: 0.8 -%}
{%- assign discount_percent = 20 -%}

<p>Sale: {{ sale_price | money }}</p>
<p>Save {{ discount_percent }}%!</p>
```

### Conditionals

```liquid
{%- if product.available? -%}
  <button>Add to Cart</button>
{%- elsif product.coming_soon? -%}
  <p>Coming Soon</p>
{%- else -%}
  <p>Out of Stock</p>
{%- endif -%}

{%- unless product.discontinued? -%}
  <p>Available for purchase</p>
{%- endunless -%}
```

### Logical Operators

```liquid
{%- if product.on_sale? and product.in_stock? -%}
  Limited time offer!
{%- endif -%}

{%- if customer.vip? or order.total > 1000 -%}
  Free shipping!
{%- endif -%}
```

### Loops

```liquid
{%- for product in collection.products -%}
  <div>{{ product.name }}</div>
{%- endfor -%}

{%- for item in cart.items limit: 5 -%}
  {{ item.name }}
{%- endfor -%}

{%- for i in (1..10) -%}
  {{ i }}
{%- endfor -%}
```

### Loop Variables

```liquid
{%- for product in products -%}
  {{ forloop.index }}      {%- comment -%} 1, 2, 3... {%- endcomment -%}
  {{ forloop.index0 }}     {%- comment -%} 0, 1, 2... {%- endcomment -%}
  {{ forloop.first }}      {%- comment -%} true on first iteration {%- endcomment -%}
  {{ forloop.last }}       {%- comment -%} true on last iteration {%- endcomment -%}
  {{ forloop.length }}     {%- comment -%} total iterations {%- endcomment -%}
{%- endfor -%}
```

---

## Template Types

### Pages

Full-page templates that define main content structure. Located in `pages/` directory.

**Example**: `pages/product.liquid`
```liquid
{%- layout "application" %}

<div class="product-page">
  <h1>{{ current_product.name }}</h1>

  <div class="product-images">
    {%- for image in current_product.images -%}
      <img src="{{ image.url }}" alt="{{ image.alt_text }}" />
    {%- endfor -%}
  </div>

  <div class="product-price">
    {{ current_product.pricing.price | money }}
  </div>

  <form action="/products/{{ current_product.id }}/add_to_cart" method="post">
    <button type="submit">Add to Cart</button>
  </form>
</div>
```

**Common Pages**:
- `home.liquid` - Homepage
- `product.liquid` - Product detail
- `products.liquid` - Product listing
- `cart.liquid` - Shopping cart
- `checkout.liquid` - Checkout process
- `order.liquid` - Order confirmation
- `account.liquid` - Customer account

### Snippets

Reusable template fragments included in pages or other snippets. Located in `snippets/` directory.

**Example**: `snippets/product_card.liquid`
```liquid
{%- comment -%} Accept parameters {%- endcomment -%}
{%- default product: nil -%}
{%- default show_compare: false -%}

{%- if product -%}
  <div class="product-card" id="product-{{ product.id }}">
    <a href="{{ product.path }}">
      {%- if product.image -%}
        <img src="{{ product.image.medium_url }}" alt="{{ product.name }}" loading="lazy" />
      {%- endif -%}

      <h3>{{ product.name }}</h3>
      <p>{{ product.pricing.price | money }}</p>
    </a>

    {%- if show_compare -%}
      <input type="checkbox" name="compare" value="{{ product.id }}" />
    {%- endif -%}
  </div>
{%- endif -%}
```

**Usage**:
```liquid
{% render "product_card", product: product, show_compare: true %}
```

### Blocks

Content blocks used by the CMS system. Located in `blocks/` directory.

**Example**: `blocks/featured_products.liquid`
```liquid
<div class="featured-products">
  {%- if content_block.title -%}
    <h2>{{ content_block.title }}</h2>
  {%- endif -%}

  <div class="product-grid">
    {%- for product in content_block.products -%}
      {% render "product_card", product: product %}
    {%- endfor -%}
  </div>
</div>
```

### Layouts

Wrapper templates for pages. Located in `layouts/` directory.

**Example**: `layouts/application.liquid`
```liquid
<!DOCTYPE html>
<html>
<head>
  <title>{{ current_store.name }}</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  {% require "styles/theme.css" %}
</head>
<body>
  {% render "header" %}

  <main>
    {{ yield }}  {%- comment -%} Page content inserted here {%- endcomment -%}
  </main>

  {% render "footer" %}

  {% require "scripts/theme.js" %}
</body>
</html>
```

**Usage in page**:
```liquid
{%- layout "application" %}

<h1>Page content here</h1>
```

### Helpers

Templates for specific functionality. Located in `helpers/` directory.

**Example**: `helpers/delivery_options.liquid`
```liquid
{%- if current_delivery_options -%}
  <div class="delivery-options">
    {%- for option in current_delivery_options.options -%}
      <label>
        <input type="radio" name="delivery" value="{{ option.id }}" />
        {{ option.name }} - {{ option.price | money }}
      </label>
    {%- endfor -%}
  </div>
{%- endif -%}
```

---

## Global Objects

Global objects are available in all templates. Access them directly without prefixes.

### Store & Configuration

#### `current_store`

The current store configuration.

```liquid
{{ current_store.name }}
{{ current_store.email }}
{{ current_store.phone }}
{{ current_store.currency }}
{{ current_store.address.full }}
```

**Properties**:
- `name` - Store name
- `email` - Contact email
- `phone` - Contact phone
- `currency` - Currency code (USD, EUR, etc.)
- `address` - Store address
- `logo` - Store logo image
- `timezone` - Store timezone

#### `theme_variables`

Theme configuration variables.

```liquid
{{ theme_variables["products.per_page"] }}
{{ theme_variables["checkout.show_delivery_date"] }}
{{ theme_variables["images.ratio.width"] }}
```

#### `store_variables`

Store-specific variables.

```liquid
{{ store_variables["custom_message"] }}
{{ store_variables["promo_banner_text"] }}
```

#### `session_variables`

Session-scoped variables.

```liquid
{{ session_variables["last_viewed_product"] }}
{{ session_variables["referral_code"] }}
```

### Customer & Account

#### `current_customer`

The logged-in customer (if authenticated).

```liquid
{%- if current_customer -%}
  <p>Welcome, {{ current_customer.first_name }}!</p>
  <p>Email: {{ current_customer.email }}</p>
  <p>Member since: {{ current_customer.created_at | date: "%B %Y" }}</p>
{%- else -%}
  <a href="/auth/login">Sign In</a>
{%- endif -%}
```

**Properties**:
- `id` - Customer ID
- `email` - Email address
- `first_name`, `last_name`, `name` - Name fields
- `phone` - Phone number
- `created_at` - Registration date
- `vip?` - VIP status
- `orders` - Customer orders
- `addresses` - Saved addresses

#### `current_account`

The customer's account.

```liquid
{{ current_account.balance | money }}
{{ current_account.points }}
{{ current_account.credit_limit | money }}
```

### Cart & Orders

#### `current_cart`

The active shopping cart.

```liquid
<p>Items: {{ current_cart.item_count }}</p>
<p>Subtotal: {{ current_cart.subtotal | money }}</p>
<p>Tax: {{ current_cart.tax | money }}</p>
<p>Total: {{ current_cart.total | money }}</p>

{%- for item in current_cart.items -%}
  <div>
    {{ item.name }} - Qty: {{ item.quantity }}
    Price: {{ item.unit_price | money }}
  </div>
{%- endfor -%}
```

**Properties**:
- `items` - Cart items collection
- `item_count` - Total item quantity
- `subtotal` - Subtotal amount
- `tax` - Tax amount
- `shipping` - Shipping cost
- `total` - Total amount
- `promotion` - Applied promotion

#### `current_order`

The current order (after checkout).

```liquid
<h1>Order #{{ current_order.order_number }}</h1>
<p>Status: {{ current_order.status }}</p>
<p>Total: {{ current_order.total | money }}</p>

{%- for item in current_order.items -%}
  {{ item.product.name }} × {{ item.quantity }}
{%- endfor -%}
```

### Products & Categories

#### `current_product`

The current product (on product pages).

```liquid
<h1>{{ current_product.name }}</h1>
<p>{{ current_product.product_code }}</p>
<p>Price: {{ current_product.pricing.price | money }}</p>

{%- if current_product.on_sale? -%}
  <span>Sale!</span>
  <p>Was: {{ current_product.pricing.list_price | money }}</p>
{%- endif -%}

{%- if current_product.in_stock? -%}
  <button>Add to Cart</button>
{%- else -%}
  <p>{{ current_product.out_of_stock_text }}</p>
{%- endif -%}

{%- for image in current_product.images -%}
  <img src="{{ image.large_url }}" alt="{{ image.alt_text }}" />
{%- endfor -%}

{%- for category in current_product.categories -%}
  <a href="{{ category.path }}">{{ category.name }}</a>
{%- endfor -%}
```

**Properties**:
- `id`, `name`, `product_code`, `barcode`
- `path` - Product URL
- `pricing` - Pricing information
- `images` - Product images
- `categories` - Product categories
- `brand` - Product brand
- `in_stock?`, `out_of_stock?`, `on_sale?` - Status checks
- `description`, `summary` - Content fields

#### `current_product_category`

The current product category.

```liquid
<h1>{{ current_product_category.name }}</h1>
<p>{{ current_product_category.description }}</p>

{%- for product in current_product_category.products -%}
  {% render "product_card", product: product %}
{%- endfor -%}
```

#### `all_products`

Lookup and paginate all products.

```liquid
{%- comment -%} Lookup by slug {%- endcomment -%}
{%- assign featured = all_products["featured-widget"] -%}
{{ featured.name }}

{%- comment -%} Pagination {%- endcomment -%}
{%- assign page = all_products.page(1, 12) -%}
{%- for product in page.results -%}
  {{ product.name }}
{%- endfor -%}
```

#### `all_product_categories`

Lookup and paginate all product categories.

```liquid
{%- assign category = all_product_categories["electronics"] -%}
{{ category.name }}
```

### Content

#### `current_page`

The current content page.

```liquid
<h1>{{ current_page.title }}</h1>
<div>{{ current_page.body_content }}</div>

<p>Published: {{ current_page.published_at | date: "%B %d, %Y" }}</p>
```

#### `current_article`

The current article/blog post.

```liquid
<article>
  <h1>{{ current_article.title }}</h1>
  <p>By {{ current_article.author.name }} on {{ current_article.publish_on | date: "%B %d, %Y" }}</p>

  {%- if current_article.image -%}
    <img src="{{ current_article.image.url }}" alt="{{ current_article.title }}" />
  {%- endif -%}

  <div>{{ current_article.body_content }}</div>

  <div>
    {%- for category in current_article.categories -%}
      <a href="{{ category.path }}">{{ category.name }}</a>
    {%- endfor -%}
  </div>
</article>
```

#### `all_pages`

Lookup and paginate all pages.

```liquid
{%- assign about = all_pages["about-us"] -%}
```

#### `all_articles`

Lookup and paginate all articles.

```liquid
{%- assign articles_page = all_articles.page(1, 10) -%}
{%- for article in articles_page.results -%}
  <h2>{{ article.title }}</h2>
{%- endfor -%}
```

### System

#### `current_request`

Information about the current HTTP request.

```liquid
{{ current_request.path }}           {%- comment -%} /products/widget {%- endcomment -%}
{{ current_request.params.page }}    {%- comment -%} Query parameter {%- endcomment -%}
{{ current_request.referrer }}       {%- comment -%} Previous URL {%- endcomment -%}
{{ current_request.user_agent }}     {%- comment -%} Browser info {%- endcomment -%}
```

#### `current_flash`

Flash messages (alerts, notices).

```liquid
{%- if current_flash.alert -%}
  <div class="alert">{{ current_flash.alert }}</div>
{%- endif -%}

{%- if current_flash.notice -%}
  <div class="notice">{{ current_flash.notice }}</div>
{%- endif -%}
```

#### `current_breadcrumbs`

Navigation breadcrumbs.

```liquid
<nav>
  {%- for crumb in current_breadcrumbs -%}
    <a href="{{ crumb.path }}">{{ crumb.name }}</a>
    {%- unless forloop.last -%} / {%- endunless -%}
  {%- endfor -%}
</nav>
```

#### `current_search`

Search results and query.

```liquid
<p>Showing results for "{{ current_search.query }}"</p>
<p>Found {{ current_search.total }} results</p>

{%- for product in current_search.products -%}
  {{ product.name }}
{%- endfor -%}
```

---

## Liquid Tags

### Control Flow

#### `{% if %}` / `{% elsif %}` / `{% else %}` / `{% endif %}`

Conditional logic.

```liquid
{%- if product.in_stock? -%}
  <button>Buy Now</button>
{%- elsif product.backorder? -%}
  <button>Pre-order</button>
{%- else -%}
  <p>Out of Stock</p>
{%- endif -%}
```

#### `{% unless %}` / `{% endunless %}`

Negative conditional (opposite of `if`).

```liquid
{%- unless product.discontinued? -%}
  <p>Available</p>
{%- endunless -%}
```

#### `{% case %}` / `{% when %}` / `{% endcase %}`

Switch statement.

```liquid
{%- case product.status -%}
{%- when "active" -%}
  In Stock
{%- when "discontinued" -%}
  No Longer Available
{%- when "coming_soon" -%}
  Coming Soon
{%- else -%}
  Contact Us
{%- endcase -%}
```

### Iteration

#### `{% for %}` / `{% endfor %}`

Loop through collections.

```liquid
{%- for product in products -%}
  {{ product.name }}
{%- endfor -%}

{%- for product in products limit: 10 offset: 5 -%}
  {{ product.name }}
{%- endfor -%}
```

#### `{% break %}` / `{% continue %}`

Control loop execution.

```liquid
{%- for product in products -%}
  {%- if product.hidden? -%}
    {%- continue -%}  {%- comment -%} Skip this iteration {%- endcomment -%}
  {%- endif -%}

  {%- if forloop.index > 10 -%}
    {%- break -%}  {%- comment -%} Exit loop {%- endcomment -%}
  {%- endif -%}

  {{ product.name }}
{%- endfor -%}
```

### Variable Assignment

#### `{% assign %}`

Create variables.

```liquid
{%- assign sale_price = product.price | times: 0.8 -%}
{%- assign formatted_date = order.created_at | date: "%B %d, %Y" -%}
```

#### `{% capture %}` / `{% endcapture %}`

Capture output into a variable.

```liquid
{%- capture product_link -%}
  <a href="{{ product.path }}">{{ product.name }}</a>
{%- endcapture -%}

{{ product_link }}
```

#### `{% default %}`

Set default values for variables (useful in snippets).

```liquid
{%- default product: nil -%}
{%- default show_price: true -%}
{%- default css_class: "product-card" -%}
```

### Template Inclusion

#### `{% render %}`

Render a snippet with parameters.

```liquid
{% render "product_card", product: product, show_compare: true %}
{% render "header/navigation" %}
{% render "cart/summary", cart: current_cart %}
```

Parameters are isolated - variables from the parent template aren't automatically available.

#### `{% include %}`

**Deprecated** - Use `{% render %}` instead.

### Caching

#### `{% cache %}` / `{% endcache %}`

Cache template fragments for performance.

```liquid
{%- cache "product-list", items: [current_store, current_customer] -%}
  {%- comment -%} Expensive rendering here {%- endcomment -%}
  {%- for product in products -%}
    {% render "product_card", product: product %}
  {%- endfor -%}
{%- endcache -%}
```

**Cache Key**: Combination of cache name and items ensures cache invalidation when data changes.

### Resources

#### `{% require %}`

Load JavaScript or CSS files.

```liquid
{%- require "scripts/product-zoom.js" -%}
{%- require "styles/checkout.css" -%}
```

Resources are fingerprinted and loaded once per page.

### Layout

#### `{% layout %}`

Specify which layout wraps the page.

```liquid
{%- layout "application" %}
{%- layout "account" %}
{%- layout "checkout" %}
```

Layouts contain `{{ yield }}` where page content is inserted.

### Comments

#### `{% comment %}` / `{% endcomment %}`

Add comments (not rendered in output).

```liquid
{%- comment -%}
  This is a comment explaining the logic below
{%- endcomment -%}
```

---

## Liquid Filters

Filters modify output. Chain multiple filters with `|`.

### String Filters

```liquid
{{ "hello world" | capitalize }}      {%- comment -%} Hello world {%- endcomment -%}
{{ "hello world" | upcase }}          {%- comment -%} HELLO WORLD {%- endcomment -%}
{{ "HELLO WORLD" | downcase }}        {%- comment -%} hello world {%- endcomment -%}
{{ "Hello World" | truncate: 8 }}     {%- comment -%} Hello... {%- endcomment -%}
{{ "  trim me  " | strip }}           {%- comment -%} trim me {%- endcomment -%}
{{ "replace-me" | replace: "-", " " }} {%- comment -%} replace me {%- endcomment -%}
{{ "one,two,three" | split: "," }}    {%- comment -%} Array: ["one", "two", "three"] {%- endcomment -%}
```

### Number Filters

```liquid
{{ 1234.56 | money }}                 {%- comment -%} $1,234.56 {%- endcomment -%}
{{ 0.15 | percentage }}               {%- comment -%} 15% {%- endcomment -%}
{{ 1234 | number }}                   {%- comment -%} 1,234 {%- endcomment -%}
{{ 1000 | points }}                   {%- comment -%} 1,000 points {%- endcomment -%}

{{ 5 | plus: 3 }}                     {%- comment -%} 8 {%- endcomment -%}
{{ 5 | minus: 2 }}                    {%- comment -%} 3 {%- endcomment -%}
{{ 5 | times: 3 }}                    {%- comment -%} 15 {%- endcomment -%}
{{ 10 | divided_by: 3 }}              {%- comment -%} 3 {%- endcomment -%}
{{ 10 | modulo: 3 }}                  {%- comment -%} 1 {%- endcomment -%}

{{ -5 | abs }}                        {%- comment -%} 5 {%- endcomment -%}
{{ 4.5 | ceil }}                      {%- comment -%} 5 {%- endcomment -%}
{{ 4.5 | floor }}                     {%- comment -%} 4 {%- endcomment -%}
{{ 4.5 | round }}                     {%- comment -%} 5 {%- endcomment -%}
```

### Date Filters

```liquid
{{ order.created_at | date: "%B %d, %Y" }}       {%- comment -%} January 14, 2026 {%- endcomment -%}
{{ order.created_at | date: "%Y-%m-%d" }}        {%- comment -%} 2026-01-14 {%- endcomment -%}
{{ order.created_at | time_ago }}                {%- comment -%} 2 hours ago {%- endcomment -%}
{{ order.created_at | format_date }}             {%- comment -%} Store's date format {%- endcomment -%}
```

### Array Filters

```liquid
{{ products | size }}                            {%- comment -%} Number of items {%- endcomment -%}
{{ products | first }}                           {%- comment -%} First product {%- endcomment -%}
{{ products | last }}                            {%- comment -%} Last product {%- endcomment -%}
{{ products | map: "name" }}                     {%- comment -%} Array of names {%- endcomment -%}
{{ products | where: "on_sale?", true }}         {%- comment -%} Filter by property {%- endcomment -%}
{{ products | sort: "name" }}                    {%- comment -%} Sort by property {%- endcomment -%}
{{ products | reverse }}                         {%- comment -%} Reverse order {%- endcomment -%}
{{ products | uniq }}                            {%- comment -%} Remove duplicates {%- endcomment -%}
{{ products | concat: more_products }}           {%- comment -%} Combine arrays {%- endcomment -%}
```

### URL Filters

```liquid
{{ "logo" | asset_url }}                         {%- comment -%} /assets/logo.png {%- endcomment -%}
{{ "search term" | url_encode }}                 {%- comment -%} search%20term {%- endcomment -%}
{{ "/cart" | link_to: "View Cart" }}             {%- comment -%} <a href="/cart">View Cart</a> {%- endcomment -%}
```

### Utility Filters

```liquid
{{ variable | default: "fallback" }}             {%- comment -%} Use fallback if nil {%- endcomment -%}
{{ object | json }}                              {%- comment -%} Convert to JSON {%- endcomment -%}
{{ "hello-world" | parameterize }}               {%- comment -%} hello-world {%- endcomment -%}
{{ html_string | escape }}                       {%- comment -%} Escape HTML {%- endcomment -%}
{{ html_string | strip_html }}                   {%- comment -%} Remove HTML tags {%- endcomment -%}
```

### Translation Filter

```liquid
{{ "products.add_to_cart" | t }}                 {%- comment -%} Translated text {%- endcomment -%}
{{ "products.price_with_name" | t: name: product.name }}  {%- comment -%} With interpolation {%- endcomment -%}
```

---

## Liquid Controllers

Liquid controllers execute at specific points in the request lifecycle, allowing you to manipulate data and control responses.

### Controller Lifecycle

Controllers execute at three stages:

1. **`{% before %}`** - Before the page loads
2. **`{% after %}`** - After the page loads (has access to page data)
3. **`{% final %}`** - At the very end

### Controller Location

Place controllers in `controllers/` directory:
- `controllers/products/show.liquid` - For specific pages
- `controllers/theme/before.liquid` - Global theme controller

### Basic Controller

```liquid
{% before %}
  {%- comment -%} Runs before page loads {%- endcomment -%}

  {%- if current_customer == blank -%}
    {%- redirect to: "/auth/login", alert: "Please sign in" %}
  {%- endif -%}
{% endbefore %}

{% after %}
  {%- comment -%} Runs after page loads {%- endcomment -%}
  {%- comment -%} Can access page variables {%- endcomment -%}
{% endafter %}

{% final %}
  {%- comment -%} Always runs last {%- endcomment -%}
{% endfinal %}
```

### Controller Actions

Execute actions using `controller.do_action`:

#### Cart Actions

```liquid
{% before %}
  {%- comment -%} Add product to cart {%- endcomment -%}
  {{ controller.do_action task: "cart.add",
     product_identifier: "PROD-123",
     quantity: 2,
     price: 99.99 }}

  {%- comment -%} Update cart item {%- endcomment -%}
  {{ controller.do_action task: "cart.update",
     cart_item: current_cart.items.first,
     quantity: 5 }}

  {%- comment -%} Remove cart item {%- endcomment -%}
  {{ controller.do_action task: "cart.remove",
     cart_item: current_cart.items.first }}

  {%- comment -%} Empty cart {%- endcomment -%}
  {{ controller.do_action task: "cart.empty" }}

  {%- comment -%} Select specific cart {%- endcomment -%}
  {{ controller.do_action task: "cart.select",
     cart_identifier: "CART-SFID" }}

  {%- comment -%} Clone a cart {%- endcomment -%}
  {{ controller.do_action task: "cart.clone",
     cart_identifier: "CART-SFID" }}
{% endbefore %}
```

#### Promotion Actions

```liquid
{% before %}
  {%- comment -%} Apply promo code {%- endcomment -%}
  {{ controller.do_action task: "promotion.apply", code: "SAVE20" }}

  {%- comment -%} Remove promo code {%- endcomment -%}
  {{ controller.do_action task: "promotion.remove", code: "SAVE20" }}

  {%- comment -%} Clear all promotions {%- endcomment -%}
  {{ controller.do_action task: "promotion.clear" }}
{% endbefore %}
```

#### Pricebook Actions

```liquid
{% before %}
  {%- comment -%} Set active pricebook {%- endcomment -%}
  {{ controller.do_action task: "pricebook.set", pricebook_id: "PB123" }}

  {%- comment -%} Clear pricebook {%- endcomment -%}
  {{ controller.do_action task: "pricebook.clear" }}
{% endbefore %}
```

#### Shipping Actions

```liquid
{% before %}
  {%- comment -%} Set shipping method {%- endcomment -%}
  {{ controller.do_action task: "shipping.set",
     price: 15.00,
     name: "Express Shipping" }}
{% endbefore %}
```

### Responding Early

Controllers can respond or redirect before the normal page renders:

```liquid
{% before %}
  {%- comment -%} Respond with custom page {%- endcomment -%}
  {%- if store_in_maintenance? -%}
    {%- respond with: "maintenance", layout: "simple", status: 503 %}
  {%- endif -%}

  {%- comment -%} Redirect to another page {%- endcomment -%}
  {%- if current_product.discontinued? -%}
    {%- redirect to: "/products", alert: "Product no longer available" %}
  {%- endif -%}

  {%- comment -%} Redirect with notice {%- endcomment -%}
  {%- redirect to: "/cart", notice: "Item added successfully" %}
{% endbefore %}
```

### Setting Variables

Pass data from controllers to views:

```liquid
{% before %}
  {%- assign featured = all_products["featured-item"] -%}
  {{ controller.set_variables featured_product: featured }}

  {%- assign special_message = "Limited Time Offer!" -%}
  {{ controller.set_variables promo_message: special_message }}
{% endbefore %}
```

Access in templates:
```liquid
{{ controller.variables.featured_product.name }}
{{ controller.variables.promo_message }}
```

### Practical Examples

#### Auto-apply promotion for VIP customers:

```liquid
{% before %}
  {%- if current_customer.vip? -%}
    {{ controller.do_action task: "promotion.apply", code: "VIP-DISCOUNT" }}
  {%- endif -%}
{% endbefore %}
```

#### Redirect if product restricted:

```liquid
{% before %}
  {%- if current_product.restricted? and current_customer == blank -%}
    {%- redirect to: "/auth/login", alert: "Please sign in to view this product" %}
  {%- endif -%}
{% endbefore %}
```

#### Set up bundle products:

```liquid
{% before %}
  {%- if current_product.is_bundle? -%}
    {%- for bundle_item in current_product.bundle_items -%}
      {{ controller.do_action task: "cart.add",
         product_identifier: bundle_item.product.slug,
         quantity: bundle_item.quantity }}
    {%- endfor -%}
  {%- endif -%}
{% endbefore %}
```

---

## Forms

StoreConnect provides pre-built forms for common operations.

### Cart Form

```liquid
<form action="/cart_items" method="post">
  <input type="hidden" name="product_id" value="{{ product.id }}" />
  <input type="number" name="quantity" value="1" min="1" />
  <button type="submit">Add to Cart</button>
</form>
```

### Update Cart Item

```liquid
<form action="/cart_items/{{ item.id }}" method="post">
  <input type="hidden" name="_method" value="patch" />
  <input type="number" name="quantity" value="{{ item.quantity }}" />
  <button type="submit">Update</button>
</form>
```

### Apply Promotion

```liquid
<form action="/carts/{{ current_cart.id }}/promotion" method="post">
  <input type="text" name="code" placeholder="Promo code" />
  <button type="submit">Apply</button>
</form>
```

### Customer Login

```liquid
<form action="/auth/session" method="post">
  <input type="email" name="email" required />
  <input type="password" name="password" required />
  <button type="submit">Sign In</button>
</form>
```

### Customer Registration

```liquid
<form action="/auth/registration" method="post">
  <input type="text" name="first_name" required />
  <input type="text" name="last_name" required />
  <input type="email" name="email" required />
  <input type="password" name="password" required />
  <button type="submit">Create Account</button>
</form>
```

### Search Form

```liquid
<form action="/search" method="get">
  <input type="text" name="query" placeholder="Search..." />
  <button type="submit">Search</button>
</form>
```

### Newsletter Signup

```liquid
<form action="/newsletter/subscribe" method="post">
  <input type="email" name="email" required />
  <button type="submit">Subscribe</button>
</form>
```

---

## Working with Data

### Pagination

```liquid
{%- assign page_number = current_request.params.page | default: 1 -%}
{%- assign per_page = 12 -%}
{%- assign products_page = all_products.page(page_number, per_page) -%}

<div class="products">
  {%- for product in products_page.results -%}
    {% render "product_card", product: product %}
  {%- endfor -%}
</div>

{%- if products_page.total_pages > 1 -%}
  <nav class="pagination">
    {%- if products_page.previous_page -%}
      <a href="?page={{ products_page.previous_page }}">Previous</a>
    {%- endif -%}

    <span>Page {{ products_page.current_page }} of {{ products_page.total_pages }}</span>

    {%- if products_page.next_page -%}
      <a href="?page={{ products_page.next_page }}">Next</a>
    {%- endif -%}
  </nav>
{%- endif -%}
```

### Lookups

```liquid
{%- comment -%} Find by slug/path {%- endcomment -%}
{%- assign product = all_products["featured-widget"] -%}
{%- assign category = all_product_categories["electronics/computers"] -%}
{%- assign page = all_pages["about-us"] -%}

{%- comment -%} Find by ID {%- endcomment -%}
{%- assign product = all_products["a1234567890ABCD"] -%}
```

### Filtering

```liquid
{%- comment -%} Filter products {%- endcomment -%}
{%- assign on_sale = products | where: "on_sale?", true -%}
{%- assign in_stock = products | where: "in_stock?", true -%}
{%- assign expensive = products | where: "price", ">", 100 -%}

{%- comment -%} Filter by brand {%- endcomment -%}
{%- assign nike_products = products | where: "brand.name", "Nike" -%}
```

### Sorting

```liquid
{%- assign sorted_by_name = products | sort: "name" -%}
{%- assign sorted_by_price = products | sort: "pricing.price" -%}
{%- assign newest_first = articles | sort: "publish_on" | reverse -%}
```

### Grouping

```liquid
{%- comment -%} Group products by category {%- endcomment -%}
{%- assign groups = products | group_by: "primary_category.name" -%}

{%- for group in groups -%}
  <h2>{{ group.name }}</h2>
  {%- for product in group.items -%}
    {{ product.name }}
  {%- endfor -%}
{%- endfor -%}
```

---

## Best Practices

### Performance

**1. Use caching for expensive operations**:
```liquid
{%- cache "homepage-featured", items: [current_store] -%}
  {%- for product in featured_products -%}
    {% render "product_card", product: product %}
  {%- endfor -%}
{%- endcache -%}
```

**2. Limit queries and loops**:
```liquid
{%- comment -%} Good {%- endcomment -%}
{%- assign products = all_products.page(1, 12) -%}

{%- comment -%} Avoid {%- endcomment -%}
{%- for product in all_products -%}  {%- comment -%} Loads everything! {%- endcomment -%}
```

**3. Lazy load images**:
```liquid
<img src="{{ product.image.url }}" loading="lazy" alt="{{ product.name }}" />
```

**4. Use snippets efficiently**:
```liquid
{%- comment -%} Don't pass entire collections {%- endcomment -%}
{% render "product_card", product: product %}

{%- comment -%} Not: {% render "product_card", all_products: all_products %} {%- endcomment -%}
```

### Code Organization

**1. Keep snippets small and focused**:
```liquid
{%- comment -%} Good: snippets/product/price.liquid {%- endcomment -%}
{%- comment -%} Good: snippets/product/image.liquid {%- endcomment -%}
{%- comment -%} Good: snippets/product/add_to_cart.liquid {%- endcomment -%}
```

**2. Use descriptive names**:
```liquid
{% render "product/add_to_cart_button" %}  {%- comment -%} Good {%- endcomment -%}
{% render "product/btn" %}                 {%- comment -%} Avoid {%- endcomment -%}
```

**3. Set defaults in snippets**:
```liquid
{%- default product: nil -%}
{%- default show_price: true -%}
{%- default button_text: "Add to Cart" -%}
```

### Security

**1. Escape user-generated content**:
```liquid
{{ user_review.comment | escape }}
{{ customer_note | strip_html }}
```

**2. Validate before using**:
```liquid
{%- if current_customer.can_purchase? -%}
  <button>Buy Now</button>
{%- endif -%}
```

### Maintainability

**1. Document complex logic**:
```liquid
{%- comment -%}
  Calculate tiered discount:
  - 0-10 items: no discount
  - 11-20 items: 10% off
  - 21+ items: 20% off
{%- endcomment -%}
{%- if cart.item_count > 20 -%}
  {%- assign discount = 0.20 -%}
{%- elsif cart.item_count > 10 -%}
  {%- assign discount = 0.10 -%}
{%- else -%}
  {%- assign discount = 0 -%}
{%- endif -%}
```

**2. Use meaningful variable names**:
```liquid
{%- assign discounted_price = product.price | times: 0.8 -%}  {%- comment -%} Good {%- endcomment -%}
{%- assign p = product.price | times: 0.8 -%}                 {%- comment -%} Avoid {%- endcomment -%}
```

**3. Avoid deep nesting**:
```liquid
{%- comment -%} Instead of deeply nested ifs... {%- endcomment -%}
{%- if condition_a -%}
  {%- if condition_b -%}
    {%- if condition_c -%}
      ...
    {%- endif -%}
  {%- endif -%}
{%- endif -%}

{%- comment -%} Use early returns or case statements {%- endcomment -%}
{%- unless condition_a -%}{% break %}{%- endunless -%}
{%- unless condition_b -%}{% break %}{%- endunless -%}
```

### Accessibility

**1. Use semantic HTML**:
```liquid
<article>
  <h1>{{ article.title }}</h1>
  <time datetime="{{ article.publish_on | date: '%Y-%m-%d' }}">
    {{ article.publish_on | date: "%B %d, %Y" }}
  </time>
</article>
```

**2. Add alt text to images**:
```liquid
<img src="{{ product.image.url }}" alt="{{ product.name }}" />
```

**3. Use ARIA labels**:
```liquid
<button aria-label="Add {{ product.name }} to cart">
  Add to Cart
</button>
```

---

## Additional Resources

- **Theme Documentation**: See `themes.md` for theme structure and development
- **Official Repositories**: https://github.com/GetStoreConnect
- **Support**: https://support.getstoreconnect.com

---

## Quick Reference

### Common Patterns

**Display product with purchase button**:
```liquid
<div class="product">
  <h1>{{ current_product.name }}</h1>
  <p>{{ current_product.pricing.price | money }}</p>

  {%- if current_product.in_stock? -%}
    <form action="/products/{{ current_product.id }}/add_to_cart" method="post">
      <input type="number" name="quantity" value="1" min="1" />
      <button type="submit">Add to Cart</button>
    </form>
  {%- else -%}
    <p>{{ current_product.out_of_stock_text }}</p>
  {%- endif -%}
</div>
```

**Loop through cart items**:
```liquid
{%- for item in current_cart.items -%}
  <div>
    <h3>{{ item.product.name }}</h3>
    <p>Qty: {{ item.quantity }}</p>
    <p>Price: {{ item.unit_price | money }}</p>
    <p>Subtotal: {{ item.total_price | money }}</p>
  </div>
{%- endfor -%}

<p>Total: {{ current_cart.total | money }}</p>
```

**Paginated product listing**:
```liquid
{%- assign page = current_request.params.page | default: 1 -%}
{%- assign products = all_products.page(page, 12) -%}

{%- for product in products.results -%}
  {% render "product_card", product: product %}
{%- endfor -%}

{%- if products.total_pages > 1 -%}
  <nav>
    {%- for i in (1..products.total_pages) -%}
      <a href="?page={{ i }}">{{ i }}</a>
    {%- endfor -%}
  </nav>
{%- endif -%}
```

**Customer account check**:
```liquid
{%- if current_customer -%}
  <p>Welcome back, {{ current_customer.first_name }}!</p>
  <a href="/account">My Account</a>
  <a href="/auth/logout">Logout</a>
{%- else -%}
  <a href="/auth/login">Sign In</a>
  <a href="/auth/register">Create Account</a>
{%- endif -%}
```

This reference covers the essential Liquid features in StoreConnect. Use it as your guide for building dynamic, data-driven storefronts.