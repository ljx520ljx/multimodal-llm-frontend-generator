import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Generate Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should show generate button disabled without images', async ({ page }) => {
    const generateButton = page.locator('button:has-text("生成交互原型")');
    await expect(generateButton).toBeDisabled();
  });

  test('should enable generate button after uploading images', async ({ page }) => {
    const testImagePath = path.join(__dirname, 'fixtures/test-images/sample.svg');

    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(testImagePath);

    const generateButton = page.locator('button:has-text("生成交互原型")');
    await expect(generateButton).toBeEnabled({ timeout: 5000 });
  });

  test('should show code editor placeholder before generation in developer mode', async ({ page }) => {
    // Switch to developer mode
    const modeToggle = page.locator('button:has-text("开发模式")');
    await modeToggle.click();

    const placeholder = page.locator('text=上传设计稿后生成代码');
    await expect(placeholder).toBeVisible();
  });

  test('should show preview placeholder before generation', async ({ page }) => {
    // The preview area should have some indication it's waiting for code
    const previewArea = page.locator('text=交互预览');
    await expect(previewArea).toBeVisible();
  });
});

test.describe('Generate Flow - Integration', () => {
  // These tests require backend to be running
  test.skip('should start generation when clicking generate button', async ({ page }) => {
    await page.goto('/');

    const testImagePath = path.join(__dirname, 'fixtures/test-images/sample.svg');
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(testImagePath);

    const generateButton = page.locator('button:has-text("生成交互原型")');
    await expect(generateButton).toBeEnabled({ timeout: 5000 });

    await generateButton.click();

    // Should show generating state
    await expect(page.locator('text=/生成中|正在/i')).toBeVisible({ timeout: 10000 });
  });
});
