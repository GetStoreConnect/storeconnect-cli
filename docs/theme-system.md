# StoreConnect Theme System

This guide explains how to build and customize themes for StoreConnect. Themes control the appearance and functionality of your storefront using the Liquid templating language.

## Table of Contents

1. [Overview](#overview)
2. [Theme Structure](#theme-structure)
3. [Template Types](#template-types)
4. [Theme Assets & Resources](#theme-assets--resources)
5. [Theme Variables & Configuration](#theme-variables--configuration)
6. [Building & Deploying Themes](#building--deploying-themes)
7. [Official StoreConnect Themes](#official-storeconnect-themes)
8. [Theme Customization](#theme-customization)
9. [Best Practices](#best-practices)

---

## Overview

StoreConnect themes are built using **Liquid**, a templating language that lets you create dynamic, data-driven storefronts. Themes consist of templates, assets, and configuration files that define your store's look and behavior.

### Key Features

- **Liquid Templating** - Access store data, products, customers, and more
- **Modular Design** - Reusable components (pages, snippets, blocks, layouts)
- **Theme Inheritance** - Custom themes build upon base templates
- **Asset Pipeline** - Modern JavaScript and CSS bundling with esbuild
- **Multi-language** - Built-in translation support
- **Flexible Configuration** - Theme variables for easy customization

### How Themes Work

1. **Requests** come into your store (e.g., `/products/widget`)
2. **Templates** render using Liquid with access to store data
3. **Layouts** wrap page content with headers/footers
4. **Assets** (CSS/JS) are loaded and applied
5. **Response** returns the complete HTML page

### Theme Resolution

When a page loads, StoreConnect looks for templates in this order:

1. **Custom Theme** - Your uploaded theme (if assigned to store)
2. **Base Theme** - Default fallback theme
3. **Built-in Theme** - System default

This means you only need to include templates you want to customize - missing templates automatically fall back to the base theme.

### Example Themes

There are

---

## Theme Structure

### Directory Layout

```
my-theme/
├── assets/                # Images, fonts (optional - can use CDN)
├── templates/             # Liquid template files
│   ├── pages/             # Full page templates
│   │   ├── home.liquid
│   │   ├── product.liquid
│   │   ├── products.liquid
│   │   ├── cart.liquid
│   │   ├── checkout.liquid
│   │   ├── order.liquid
│   │   └── account.liquid
│   ├── snippets/          # Reusable components
│   │   ├── header.liquid
│   │   ├── footer.liquid
│   │   ├── product_card.liquid
│   │   └── cart_item.liquid
│   ├── blocks/            # CMS content blocks
│   │   ├── featured_products.liquid
│   │   ├── hero.liquid
│   │   └── image_text.liquid
│   ├── layouts/           # Page wrappers
│   │   ├── application.liquid
│   │   └── account.liquid
│   ├── helpers/           # Utility templates
│   │   └── delivery_options.liquid
│   └── controllers/       # Request lifecycle hooks
│       └── products/
│           └── show.liquid
├── translations/          # Multi-language support
│   ├── en.csv
│   └── fr.csv
├── variables.csv          # Theme configuration
├── controllers.csv        # Controller mappings (advanced)
└── README.md              # Theme documentation
```

### Minimal Theme

A minimal theme can be as simple as:

```
minimal-theme/
├── templates/
│   └── pages/
│       └── home.liquid
└── variables.csv
```

All other templates fall back to the base theme.

---

## Template Types

For detailed Liquid syntax and examples, see **[Liquid Reference](liquid.md)**.

### 1. Pages

Full-page templates that define main content. Located in `templates/pages/`.

**Example**: `pages/product.liquid`
```liquid
{%- layout "application" %}

<div class="product-page">
  <h1>{{ current_product.name }}</h1>
  <p>{{ current_product.pricing.price | money }}</p>

  {% render "product/images", product: current_product %}
  {% render "product/add_to_cart", product: current_product %}
</div>
```

**Common Pages**:
- `home.liquid` - Homepage
- `product.liquid` - Product detail
- `products.liquid` - Product listing
- `cart.liquid` - Shopping cart
- `checkout.liquid` - Checkout flow
- `order.liquid` - Order confirmation
- `account.liquid` - Customer account
- `search.liquid` - Search results
- `not_found.liquid` - 404 page

### 2. Snippets

Reusable components included in pages or other snippets. Located in `templates/snippets/`.

**Example**: `snippets/product_card.liquid`
```liquid
{%- default product: nil -%}

{%- if product -%}
  <div class="product-card">
    <a href="{{ product.path }}">
      <img src="{{ product.image.medium_url }}" alt="{{ product.name }}" loading="lazy" />
      <h3>{{ product.name }}</h3>
      <p>{{ product.pricing.price | money }}</p>
    </a>
  </div>
{%- endif -%}
```

**Usage**:
```liquid
{% render "product_card", product: product %}
```

**Organization Tips**:
- Group related snippets in subdirectories (`product/`, `cart/`, `account/`)
- Keep snippets focused on single responsibility
- Use descriptive names (`product_card.liquid` not `card.liquid`)

### 3. Blocks

CMS content blocks for page builder. Located in `templates/blocks/`.

**Example**: `blocks/featured_products.liquid`
```liquid
<section class="featured-products">
  {%- if content_block.title -%}
    <h2>{{ content_block.title }}</h2>
  {%- endif -%}

  <div class="product-grid">
    {%- for product in content_block.products -%}
      {% render "product_card", product: product %}
    {%- endfor -%}
  </div>
</section>
```

### 4. Layouts

Wrapper templates that contain common elements (header, footer). Located in `templates/layouts/`.

**Example**: `layouts/application.liquid`
```liquid
<!DOCTYPE html>
<html lang="{{ current_store.locale }}">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ current_store.name }}</title>
  {% require "styles/theme.css" %}
</head>
<body>
  {% render "header" %}

  <main>
    {{ yield }}  {# Page content inserted here #}
  </main>

  {% render "footer" %}

  {% require "scripts/theme.js" %}
</body>
</html>
```

**Usage in page**:
```liquid
{%- layout "application" %}
```

### 5. Controllers

Hooks that execute at specific request lifecycle points. Located in `templates/controllers/`. See [Liquid Reference - Controllers](liquid.md#liquid-controllers) for details.

**Example**: `controllers/products/show.liquid`
```liquid
{% before %}
  {# Runs before page loads #}
  {%- if current_product.restricted? and current_customer == blank -%}
    {%- redirect to: "/auth/login", alert: "Please sign in" %}
  {%- endif -%}
{% endbefore %}
```

---

## Theme Assets & Resources

### Asset Types

1. **External Assets** - Images, fonts hosted on CDN (referenced via `assets/`)
2. **Theme Resources** - JavaScript and CSS built from source

### Resources Structure

```
resources/
├── src/
│   ├── scripts/          # JavaScript source
│   │   ├── packs/       # Entry points
│   │   └── theme/       # Shared modules
│   ├── styles/          # SCSS source
│   │   ├── packs/       # Entry points
│   │   └── theme/       # Shared styles
│   └── files/           # Static files (icons, etc.)
├── dist/                # Built output (generated)
│   ├── manifest.json    # Asset fingerprints
│   ├── scripts/
│   └── styles/
├── build/               # Build scripts
│   ├── index.js
│   ├── scripts-config.js
│   └── styles-config.js
└── package.json
```

### Building Resources

Theme resources use **esbuild** for fast bundling.

**Setup**:
```bash
cd resources/
npm install
```

**Development** (watch mode):
```bash
npm run watch
```

**Production Build**:
```bash
npm run build
```

**Clean and Rebuild**:
```bash
npm run clean
```

### Using Resources in Templates

```liquid
{%- require "scripts/product-zoom.js" -%}
{%- require "styles/checkout.css" -%}
```

Resources are:
- Automatically fingerprinted for cache-busting
- Loaded once per page (duplicates ignored)
- Minified in production builds

### Asset Manifest

Built resources generate `dist/manifest.json`:

```json
{
  "scripts/theme.js": "scripts/theme.abc123.js",
  "styles/theme.css": "styles/theme.def456.css"
}
```

This enables cache invalidation when files change.

---

## Theme Variables & Configuration

### variables.csv

Configure theme behavior without code changes.

**Format**:
```csv
Key,Value
products.per_page,12
products.comparisons,false
products.card.hide_purchase_button,false
checkout.delivery_windows.use_days,false
images.ratio.height,5
images.ratio.width,4
```

### Using Variables in Templates

```liquid
{%- assign per_page = theme_variables["products.per_page"] | default: 12 -%}
{%- assign show_comparisons = theme_variables["products.comparisons"] -%}

{%- if show_comparisons -%}
  <input type="checkbox" name="compare" />
{%- endif -%}
```

### Common Variables

**Products**:
- `products.per_page` - Products per page (default: 12)
- `products.comparisons` - Enable product comparison (default: false)
- `products.brands` - Show brand filtering (default: true)
- `products.card.hide_purchase_button` - Hide add-to-cart on cards

**Checkout**:
- `checkout.delivery_windows.use_days` - Use day-based delivery windows
- `checkout.pay_by_account.require_po_number` - Require PO for account payment

**Layout**:
- `images.ratio.height` - Default image height ratio
- `images.ratio.width` - Default image width ratio

### assets.json

Map asset keys to URLs (for CDN assets).

```json
{
  "logo": "https://cdn.example.com/logo.png",
  "banner": "https://cdn.example.com/banner.jpg",
  "icon-cart": "https://cdn.example.com/icons/cart.svg"
}
```

**Usage**:
```liquid
<img src="{{ 'logo' | asset_url }}" alt="{{ current_store.name }}" />
```

---

## Building & Deploying Themes

### Development Workflow

**1. Start with Base Theme**

Clone the base theme as your starting point:

```bash
git clone https://github.com/GetStoreConnect/base-theme my-custom-theme
cd my-custom-theme
```

**2. Set Up Resources**

```bash
cd resources/
npm install
npm run watch  # Starts development server
```

**3. Edit Templates**

Modify templates in `templates/` directory. Only include templates you want to customize.

**4. Test Locally**

- Build resources: `npm run build`
- Create ZIP of theme folder
- Upload to StoreConnect via Theme Importer

**5. Preview Theme**

Before activating, preview with: `?theme-preview=THEME_ID`

### Production Deployment

**Build Resources**:
```bash
cd resources/
npm run build  # Minified production build
```

**Package Theme**:
```bash
zip -r my-theme.zip templates/ translations/ resources/dist/ variables.csv assets.json README.md
```

**Upload**:
1. Go to StoreConnect admin
2. Navigate to Themes section
3. Click "Import Theme"
4. Upload `my-theme.zip`
5. Assign to store

### Theme Updates

When updating a theme:

1. Make changes to source files
2. Build resources (`npm run build`)
3. Re-package and upload
4. Existing customizations are preserved

---

## Official StoreConnect Themes

StoreConnect provides official themes for different use cases. All themes are available on [GitHub](https://github.com/GetStoreConnect).

### Theme Overview

| Theme | Best For | Key Features |
|-------|----------|--------------|
| **Base** | Starting point | Complete template set, foundation for all themes |
| **Clean** | Retail/Fashion | Minimalist design, product sliders, sticky header |
| **Elegance** | Fashion/Beauty | Sophisticated layout, upsell focus, image zoom |
| **Corporate** | B2B/Services | Professional design, full-width heroes, minimalist |
| **Simple Donations** | Non-profits | Donation forms, progress bars, campaign tracking |

### Base Theme

**Repository**: https://github.com/GetStoreConnect/base-theme

The default theme that all other themes inherit from. Contains complete templates for all pages, snippets, blocks, and layouts.

**When to Use**:
- Starting a new theme from scratch
- Learning theme structure
- Need complete control over all templates

**Key Files**:
- `variables.csv` - 21 configuration options
- `controllers.csv` - Controller mappings
- Complete `templates/` directory

### Clean Theme

**Repository**: https://github.com/GetStoreConnect/clean-theme

Multi-purpose theme with minimalist design.

**Features**:
- Jost font for modern typography
- Card carousels for categories/products
- Sticky header navigation
- Newsletter signup forms
- Accordion content sections

**New Blocks**:
- `accordion_tab` - Expandable sections
- `container_skinny` - Narrow containers (80% width)
- `filtered_products` - Product filtering
- `single_product_card` - Product spotlight

### Elegance Theme

**Repository**: https://github.com/GetStoreConnect/elegance-theme

Sophisticated theme for fashion and cosmetics.

**Features**:
- Upsell-optimized (add-to-cart modal shows suggestions)
- Image swapper (hover for alternate views)
- Image zoom on product cards
- Cart sidebar with quick actions
- Expandable search interface

**Best For**: Visual merchandising, high-margin products, upselling

### Corporate Theme

**Repository**: https://github.com/GetStoreConnect/corporate-theme

Professional theme for B2B and services.

**Features**:
- Full-width hero banners
- Minimalist header
- Professional footer
- Animated image overlays
- Product/service sliders

**New Blocks**:
- `hero` - Full-width banner (requires manual picklist setup)

**Best For**: B2B stores, service companies, professional brands

### Simple Donations Theme

**Repository**: https://github.com/GetStoreConnect/simple-donations-theme

Specialized theme for nonprofit organizations.

**Features**:
- Donation box with variable pricing
- Progress bars showing fundraising goals
- Campaign cards
- Testimonial sliders
- Flexible contribution amounts

**Setup Requirements**:
- Custom fields: `Donations_Total__c`, `Donation_Target__c`
- Manual template additions: `accordion`, `donation_card`, `donation_progress_bar`, `testimonial_slider`

**Best For**: Nonprofits, charities, fundraising campaigns

### Installing Official Themes

**Step 1: Download**
```bash
# Clone or download ZIP from GitHub
git clone https://github.com/GetStoreConnect/[theme-name]
```

**Step 2: Build Resources**
```bash
cd [theme-name]/resources/
npm install
npm run build
```

**Step 3: Package**
```bash
zip -r theme.zip templates/ translations/ resources/dist/ variables.csv
```

**Step 4: Upload**
- Go to StoreConnect admin → Themes
- Click "Import Theme"
- Upload `theme.zip`

**Step 5: Configure**
- Add required templates to picklists (check theme README)
- Set up custom fields if needed
- Configure theme variables

**Step 6: Activate**
- Preview: `?theme-preview=THEME_ID`
- Activate: Assign theme to store

---

## Theme Customization

### Customization Approaches

**1. Override Specific Templates**

Create a new theme with only the templates you want to change:

```
my-custom-theme/
├── templates/
│   ├── snippets/
│   │   └── header.liquid      # Custom header
│   └── pages/
│       └── home.liquid         # Custom homepage
└── variables.csv               # Custom configuration
```

All other templates fall back to base theme.

**2. Modify Existing Theme**

Clone an official theme and modify:

```bash
git clone https://github.com/GetStoreConnect/clean-theme
cd clean-theme
# Make your changes
npm run build
```

**3. Build From Base**

Start with base theme and add your customizations:

```bash
git clone https://github.com/GetStoreConnect/base-theme my-theme
cd my-theme
# Customize as needed
```

### Common Customizations

**Change Header**:

Create `templates/snippets/header.liquid`:

```liquid
<header>
  <a href="/">
    <img src="{{ 'logo' | asset_url }}" alt="{{ current_store.name }}" />
  </a>

  <nav>
    <a href="/products">Shop</a>
    <a href="/about">About</a>
    <a href="/cart">Cart ({{ current_cart.item_count }})</a>
  </nav>
</header>
```

**Customize Product Card**:

Create `templates/snippets/product_card.liquid`:

```liquid
{%- default product: nil -%}

<div class="my-product-card">
  <a href="{{ product.path }}">
    {%- if product.on_sale? -%}
      <span class="sale-badge">Sale!</span>
    {%- endif -%}

    <img src="{{ product.image.medium_url }}" alt="{{ product.name }}" />
    <h3>{{ product.name }}</h3>

    {%- if product.on_sale? -%}
      <p>
        <s>{{ product.pricing.list_price | money }}</s>
        <strong>{{ product.pricing.price | money }}</strong>
      </p>
    {%- else -%}
      <p>{{ product.pricing.price | money }}</p>
    {%- endif -%}
  </a>
</div>
```

**Add Custom Styles**:

In `resources/src/styles/packs/theme.scss`:

```scss
.my-product-card {
  border: 1px solid #e0e0e0;
  padding: 1rem;
  border-radius: 8px;

  .sale-badge {
    background: red;
    color: white;
    padding: 0.25rem 0.5rem;
    font-size: 0.875rem;
  }

  img {
    width: 100%;
    height: auto;
  }
}
```

Build and deploy:
```bash
npm run build
# Package and upload theme
```

### Translation Support

**Create Translation File**:

`translations/en.csv`:
```csv
Key,Value
products.add_to_cart,Add to Cart
products.out_of_stock,Out of Stock
cart.empty,Your cart is empty
checkout.payment,Payment Information
```

**Use in Templates**:
```liquid
<button>{{ "products.add_to_cart" | t }}</button>
<p>{{ "cart.total" | t }}: {{ current_cart.total | money }}</p>
```

**With Interpolation**:
```csv
Key,Value
products.price_with_tax,Price includes {{ tax_rate }}% tax
```

```liquid
{{ "products.price_with_tax" | t: tax_rate: 10 }}
```

---

## Best Practices

### Theme Development

**1. Start Small**

Begin with base theme, only override what you need:

```
minimal-override/
└── templates/
    ├── snippets/
    │   └── header.liquid    # Only custom header
    └── variables.csv         # Configuration
```

**2. Use Semantic Structure**

```liquid
{# Good #}
<article class="product">
  <h1 class="product__title">{{ product.name }}</h1>
  <div class="product__price">{{ product.price | money }}</div>
</article>

{# Avoid #}
<div class="box">
  <div class="text">{{ product.name }}</div>
  <div>{{ product.price | money }}</div>
</div>
```

**3. Organize Templates**

```
snippets/
├── layout/
│   ├── header.liquid
│   └── footer.liquid
├── product/
│   ├── card.liquid
│   ├── price.liquid
│   └── add_to_cart.liquid
└── cart/
    ├── item.liquid
    └── summary.liquid
```

**4. Keep Snippets Focused**

```liquid
{# Good: focused snippet #}
{% render "product/price", product: product %}

{# Avoid: monolithic snippet #}
{% render "product_everything", product: product, cart: current_cart, customer: current_customer %}
```

### Performance

**1. Cache Expensive Operations**

```liquid
{%- cache "homepage-products", items: [current_store] -%}
  {%- assign products = all_products.page(1, 12) -%}
  {%- for product in products.results -%}
    {% render "product_card", product: product %}
  {%- endfor -%}
{%- endcache -%}
```

**2. Optimize Images**

```liquid
{# Use appropriate size #}
<img src="{{ product.image.medium_url }}" alt="{{ product.name }}" loading="lazy" />

{# Avoid loading full-size unnecessarily #}
<img src="{{ product.image.url }}" />  {# Could be huge! #}
```

**3. Limit Queries**

```liquid
{# Good: paginate #}
{%- assign products = all_products.page(1, 12) -%}

{# Avoid: load everything #}
{%- for product in all_products -%}  {# Loads all products! #}
```

**4. Minimize Resource Loading**

```liquid
{# Load once in layout #}
{% require "scripts/theme.js" %}

{# Not: multiple times in snippets #}
```

### Accessibility

**1. Semantic HTML**

```liquid
<nav aria-label="Main navigation">
  <a href="/">Home</a>
  <a href="/products">Shop</a>
</nav>

<main>
  <article>
    <h1>{{ current_product.name }}</h1>
  </article>
</main>
```

**2. Alt Text**

```liquid
<img src="{{ product.image.url }}" alt="{{ product.name }}" />
<img src="{{ 'icon-cart' | asset_url }}" alt="Shopping cart" />
```

**3. Form Labels**

```liquid
<label for="email">Email Address</label>
<input type="email" id="email" name="email" required />

<button type="submit" aria-label="Add {{ product.name }} to cart">
  Add to Cart
</button>
```

### Maintenance

**1. Document Custom Logic**

```liquid
{%- comment -%}
  Tiered pricing:
  - Bulk orders (20+ items): 20% off
  - Standard orders (10-19 items): 10% off
  - Small orders (< 10 items): no discount
{%- endcomment -%}
{%- if current_cart.item_count >= 20 -%}
  {%- assign discount = 0.20 -%}
{%- elsif current_cart.item_count >= 10 -%}
  {%- assign discount = 0.10 -%}
{%- else -%}
  {%- assign discount = 0 -%}
{%- endif -%}
```

**2. Use Descriptive Names**

```liquid
{# Good #}
{%- assign sale_price = product.price | times: 0.8 -%}
{%- assign shipping_threshold = 50 -%}

{# Avoid #}
{%- assign p = product.price | times: 0.8 -%}
{%- assign t = 50 -%}
```

**3. Version Control**

```bash
git init
git add .
git commit -m "Initial theme"
git tag v1.0.0
```

**4. Keep README Updated**

```markdown
# My Custom Theme

## Features
- Custom header with mega menu
- Enhanced product cards
- Mobile-optimized checkout

## Installation
1. Build resources: `npm run build`
2. Package: `zip -r theme.zip templates/ resources/dist/`
3. Upload via Theme Importer

## Configuration
- `products.per_page`: Products per page (default: 16)
- `theme.show_brands`: Show brand filter (default: true)
```

### Testing

**1. Test Different Data States**

- Empty cart
- Full cart (many items)
- Out of stock products
- Products without images
- Long product names
- Missing customer data

**2. Test Responsive Design**

- Mobile (320px - 768px)
- Tablet (768px - 1024px)
- Desktop (1024px+)

**3. Test Browser Compatibility**

- Chrome
- Firefox
- Safari
- Edge

**4. Test Performance**

- Check page load times
- Monitor resource sizes
- Test with slow connections

---

## Additional Resources

- **Liquid Reference**: See [liquid.md](liquid.md) for complete Liquid documentation
- **GitHub Themes**: https://github.com/GetStoreConnect
- **Support**: https://support.getstoreconnect.com
- **Community**: Connect with other theme developers

---

## Quick Start Checklist

- [ ] Clone base or official theme
- [ ] Install dependencies (`npm install`)
- [ ] Start development build (`npm run watch`)
- [ ] Customize templates as needed
- [ ] Build for production (`npm run build`)
- [ ] Package theme (`zip -r theme.zip ...`)
- [ ] Upload via Theme Importer
- [ ] Preview before activating (`?theme-preview=ID`)
- [ ] Configure variables in admin
- [ ] Activate theme

Happy theming! 🎨