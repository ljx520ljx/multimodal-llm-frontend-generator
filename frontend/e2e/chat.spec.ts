import { test, expect } from '@playwright/test';

test.describe('Chat Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should not show chat panel without session', async ({ page }) => {
    // Chat panel should not be visible before generation
    const chatInput = page.locator('textarea[placeholder*="输入"]');
    await expect(chatInput).not.toBeVisible();
  });
});

test.describe('Chat Flow - With Session', () => {
  // These tests require a session with generated code
  test.skip('should show chat input after code generation', async ({ page }) => {
    // This would require mocking or actual backend
    await page.goto('/');

    // After code generation, chat panel should be visible
    const chatInput = page.locator('textarea[placeholder*="输入"]');
    await expect(chatInput).toBeVisible();
  });

  test.skip('should send chat message', async ({ page }) => {
    await page.goto('/');

    const chatInput = page.locator('textarea[placeholder*="输入"]');
    await chatInput.fill('把按钮改成蓝色');

    const sendButton = page.locator('button:has-text("发送")');
    await sendButton.click();

    // Should show user message
    await expect(page.locator('text=把按钮改成蓝色')).toBeVisible();
  });

  test.skip('should show assistant response', async ({ page }) => {
    await page.goto('/');

    const chatInput = page.locator('textarea[placeholder*="输入"]');
    await chatInput.fill('把按钮改成蓝色');

    const sendButton = page.locator('button:has-text("发送")');
    await sendButton.click();

    // Should show loading indicator
    await expect(page.locator('[class*="animate"]')).toBeVisible();
  });
});
