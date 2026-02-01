from playwright.sync_api import sync_playwright
import time
import pyotp

def verify_stock_suggestions():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()
        page = context.new_page()

        base_url = "http://localhost:3000"

        try:
            # 1. Login
            print("Navigating to Login page...")
            page.goto(f"{base_url}/login")
            page.wait_for_load_state("networkidle")

            print("Filling login credentials...")
            page.fill("input[type='email']", "test@example.com")
            page.fill("input[type='password']", "password123")

            print("Submitting login form...")
            page.click("button[type='submit']")

            # 2. OTP Step
            print("Waiting for OTP input...")
            # Wait for the OTP input to appear (it has a placeholder "000000" or similar, or we can look for the input type/class)
            # In LoginPage.tsx: input value={otpCode} ... placeholder="000000"
            page.wait_for_selector("input[placeholder='000000']", timeout=5000)

            print("Generating OTP...")
            totp = pyotp.TOTP("SECRET")
            otp_code = totp.now()
            print(f"Entering OTP: {otp_code}")

            page.fill("input[placeholder='000000']", otp_code)

            print("Submitting OTP...")
            # The button text is "CONFIRM IDENTITY"
            page.click("button:has-text('CONFIRM IDENTITY')")

            # 3. Wait for redirect
            print("Waiting for login redirect...")
            page.wait_for_url(lambda url: "/login" not in url, timeout=10000)
            print(f"Logged in. Current URL: {page.url}")

            # 4. Navigate to Stock Suggestions
            print("Navigating to Stock Suggestions page...")
            page.goto(f"{base_url}/stock-suggestions")

            # Wait for content
            print("Waiting for 'AI Stock Suggestions' header...")
            try:
                page.wait_for_selector("text=AI Stock Suggestions", timeout=15000)
                print("Found 'AI Stock Suggestions' header.")
            except Exception as e:
                print(f"Could not find header. Taking debug screenshot. Error: {e}")
                page.screenshot(path="verification/debug_suggestions_missing.png")
                raise e

            # Take verification screenshot
            print("Taking screenshot...")
            page.screenshot(path="verification/stock_suggestions.png")
            print("Screenshot saved to verification/stock_suggestions.png")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
            raise e
        finally:
            browser.close()

if __name__ == "__main__":
    verify_stock_suggestions()
