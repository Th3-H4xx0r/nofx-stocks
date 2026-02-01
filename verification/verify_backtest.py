from playwright.sync_api import sync_playwright, expect
import time
import json

def verify_backtest(page):
    print("Navigating to home...")
    page.goto("http://localhost:3009")

    print("Mocking auth...")
    page.evaluate("localStorage.setItem('auth_token', 'mock_token')")
    page.evaluate("localStorage.setItem('auth_user', JSON.stringify({id: 'mock', email: 'mock@test.com'}))")

    print("Navigating to /backtest...")
    page.goto("http://localhost:3009/backtest")

    # Wait for the page content to load
    print("Waiting for heading...")
    # The heading text is t('title') which is likely "Backtest Strategy" or similar
    # Let's wait for the H1
    expect(page.locator("h1")).to_be_visible(timeout=10000)

    # Take initial screenshot
    page.screenshot(path="verification/backtest_initial.png")

    print("Finding Asset Class selector...")
    # Find the select element that has option "Crypto"
    select = page.locator('select:has(option[value="crypto"])')
    expect(select).to_be_visible()

    print("Changing to Stock...")
    select.select_option("stock")

    # Verify symbols textarea changed
    print("Verifying symbols...")
    textarea = page.locator("textarea")
    # Wait for value to change
    expect(textarea).to_have_value("AAPL,TSLA,NVDA,MSFT,AMZN")

    print("Taking final screenshot...")
    page.screenshot(path="verification/backtest_stock.png")
    print("Verification successful!")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_backtest(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
            raise e
        finally:
            browser.close()
