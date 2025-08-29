#!/usr/bin/env python3
import requests
import json
import sys

def check_security_headers(url):
    try:
        response = requests.get(url, timeout=10)
        headers = response.headers
        
        security_headers = {
            'X-Content-Type-Options': 'nosniff',
            'X-Frame-Options': 'DENY',
            'X-XSS-Protection': '1; mode=block',
            'Strict-Transport-Security': 'max-age=31536000',
            'Content-Security-Policy': 'default-src \'self\'',
            'Referrer-Policy': 'strict-origin-when-cross-origin'
        }
        
        results = {}
        for header, expected_value in security_headers.items():
            if header in headers:
                results[header] = {
                    'present': True,
                    'value': headers[header],
                    'secure': expected_value in headers[header]
                }
            else:
                results[header] = {
                    'present': False,
                    'value': None,
                    'secure': False
                }
        
        return results
    except Exception as e:
        print(f"Error checking {url}: {e}")
        return None

if __name__ == "__main__":
    urls = [
        "http://api-gateway:8090",
        "http://database-service:8115",
        "http://monitoring-service:8116"
    ]
    
    all_results = {}
    for url in urls:
        print(f"Checking security headers for {url}")
        results = check_security_headers(url)
        if results:
            all_results[url] = results
    
    with open('/app/security-headers-report.json', 'w') as f:
        json.dump(all_results, f, indent=2)
    
    print("Security headers check completed")
