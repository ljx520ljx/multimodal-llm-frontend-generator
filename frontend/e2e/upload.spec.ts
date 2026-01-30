import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Upload Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display upload dropzone', async ({ page }) => {
    const dropzone = page.locator('text=拖拽或点击上传设计稿');
    await expect(dropzone).toBeVisible();
  });

  test('should show supported formats hint', async ({ page }) => {
    const hint = page.locator('text=/支持.*PNG.*JPG/i');
    await expect(hint).toBeVisible();
  });

  test('should upload image via file input', async ({ page }) => {
    // Create a test image file
    const testImagePath = path.join(__dirname, 'fixtures/test-images/sample.svg');

    // Find file input and upload
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(testImagePath);

    // Wait for image to appear in list
    await expect(page.locator('text=sample.svg')).toBeVisible({ timeout: 5000 });
  });

  test('should show image preview after upload', async ({ page }) => {
    const testImagePath = path.join(__dirname, 'fixtures/test-images/sample.svg');

    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(testImagePath);

    // Check that image preview is shown
    const imagePreview = page.locator('img[alt*="Image"]');
    await expect(imagePreview).toBeVisible({ timeout: 5000 });
  });

  test('should remove image from list', async ({ page }) => {
    const testImagePath = path.join(__dirname, 'fixtures/test-images/sample.svg');

    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(testImagePath);

    // Wait for image to appear
    await expect(page.locator('text=sample.svg')).toBeVisible({ timeout: 5000 });

    // Find and click delete button
    const deleteButton = page.locator('button').filter({ has: page.locator('svg path[d*="M6 18L18 6"]') });
    await deleteButton.click();

    // Verify image is removed
    await expect(page.locator('text=sample.svg')).not.toBeVisible();
  });

  test('should show generate button when images uploaded', async ({ page }) => {
    const testImagePath = path.join(__dirname, 'fixtures/test-images/sample.svg');

    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(testImagePath);

    // Wait for generate button to be enabled
    const generateButton = page.locator('button:has-text("生成交互原型")');
    await expect(generateButton).toBeVisible({ timeout: 5000 });
  });
});
