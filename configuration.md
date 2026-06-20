# Configuration Guide

This document describes how to configure domain-list-custom using the JSON configuration file.

## Configuration File Structure

The configuration file consists of two main sections: `input` and `output`.

```json
{
  "input": [...],
  "output": [...]
}
```

## Input Configuration

### Domain List Input

Type: `domainlist`

Load domain lists from a directory.

```json
{
  "type": "domainlist",
  "action": "add",
  "args": {
    "dataDir": "./data",
    "wantedList": []
  }
}
```

**Arguments:**
- `dataDir` (required): Path to the directory containing domain list files
- `wantedList` (optional): Array of specific domain lists to load. If empty, all lists are loaded.

**Domain List File Format:**

```
# Comments start with #
domain.com                    # Domain without prefix = domain type
full:exact.domain.com         # Full match
keyword:ads                   # Keyword match
regexp:.*tracker.*            # Regular expression match

# Attributes
domain.com @ads               # Domain with single attribute
domain.com @ads @cn           # Domain with multiple attributes

# Inclusions
include:other-list            # Include all domains from other-list
include:other-list @cn        # Include only domains with @cn attribute from other-list
```

## Output Configuration

### V2Ray GeoSite Output

Type: `v2rayGeoSite`

Generate V2Ray geosite.dat file and optionally GFWList.

```json
{
  "type": "v2rayGeoSite",
  "action": "output",
  "args": {
    "outputDir": "./output",
    "outputName": "geosite.dat",
    "wantedList": [],
    "excludedList": [],
    "excludeAttrs": "cn@!cn@ads,geolocation-cn@!cn@ads",
    "gfwlistOutput": "geolocation-!cn"
  }
}
```

**Arguments:**
- `outputDir` (optional): Output directory path. Default: `./output`
- `outputName` (optional): Output filename. Default: `geosite.dat`
- `wantedList` (optional): Array of lists to include. If empty, all lists are included.
- `excludedList` (optional): Array of lists to exclude.
- `excludeAttrs` (optional): Rules to exclude domains with specific attributes from specific lists. Format: `list@attr1@attr2,list2@attr3`
- `gfwlistOutput` (optional): Name of the list to generate as GFWList format.

**Exclude Attributes Format:**

To exclude domains from `cn` list that have `!cn` or `ads` attributes:
```
cn@!cn@ads
```

Multiple lists:
```
cn@!cn@ads,geolocation-cn@!cn@ads,geolocation-!cn@cn@ads
```

### Text Output

Type: `text`

Generate plaintext domain list files.

```json
{
  "type": "text",
  "action": "output",
  "args": {
    "outputDir": "./output",
    "wantedList": ["cn", "google", "apple"],
    "excludedList": []
  }
}
```

**Arguments:**
- `outputDir` (optional): Output directory path. Default: `./output`
- `wantedList` (optional): Array of lists to export. If empty, all lists are exported.
- `excludedList` (optional): Array of lists to exclude.

**Output Format:**

```
domain:example.com
full:exact.domain.com
keyword:ads
regexp:.*tracker.*
domain:domain.com:@ads
domain:domain.com:@ads,@cn
```

## Complete Example

```json
{
  "input": [
    {
      "type": "domainlist",
      "action": "add",
      "args": {
        "dataDir": "./data"
      }
    }
  ],
  "output": [
    {
      "type": "v2rayGeoSite",
      "action": "output",
      "args": {
        "outputDir": "./output",
        "outputName": "geosite.dat",
        "excludeAttrs": "cn@!cn@ads,geolocation-cn@!cn@ads,geolocation-!cn@cn@ads",
        "gfwlistOutput": "geolocation-!cn"
      }
    },
    {
      "type": "text",
      "action": "output",
      "args": {
        "outputDir": "./output",
        "wantedList": [
          "category-ads-all",
          "cn",
          "geolocation-cn",
          "geolocation-!cn",
          "google",
          "apple"
        ]
      }
    }
  ]
}
```

## Usage

```bash
# Use config.json in current directory
./domain-list-custom convert

# Use specific config file
./domain-list-custom convert -c custom-config.json

# Use remote config file
./domain-list-custom convert -c https://example.com/config.json

# List available domain lists
./domain-list-custom list -c config.json
```

## Advanced Examples

### Multiple Data Sources

You can have multiple input configurations to load data from different sources:

```json
{
  "input": [
    {
      "type": "domainlist",
      "action": "add",
      "args": {
        "dataDir": "./official-data"
      }
    },
    {
      "type": "domainlist",
      "action": "add",
      "args": {
        "dataDir": "./custom-data",
        "wantedList": ["custom-list-1", "custom-list-2"]
      }
    }
  ],
  "output": [...]
}
```

### Selective Export

Export only specific lists to different formats:

```json
{
  "input": [...],
  "output": [
    {
      "type": "v2rayGeoSite",
      "action": "output",
      "args": {
        "outputDir": "./output",
        "outputName": "geosite-cn.dat",
        "wantedList": ["cn", "geolocation-cn"]
      }
    },
    {
      "type": "v2rayGeoSite",
      "action": "output",
      "args": {
        "outputDir": "./output",
        "outputName": "geosite-ads.dat",
        "wantedList": ["category-ads-all"]
      }
    }
  ]
}
```
