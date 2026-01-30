import { describe, it, expect, beforeEach } from 'vitest';
import { useProjectStore } from '@/stores/useProjectStore';
import type { ImageFile } from '@/types';

describe('useProjectStore', () => {
  beforeEach(() => {
    useProjectStore.setState({
      codeExpanded: false,
      sessionId: null,
      images: [],
      imageIds: [],
      generatedCode: null,
      generatedCss: '',
      thinkingContent: '',
      status: 'idle',
      errorMessage: null,
      activeFile: 'App.tsx',
      chatMessages: [],
    });
  });

  describe('Code Panel', () => {
    it('should toggle code expanded', () => {
      expect(useProjectStore.getState().codeExpanded).toBe(false);
      useProjectStore.getState().toggleCodeExpanded();
      expect(useProjectStore.getState().codeExpanded).toBe(true);
      useProjectStore.getState().toggleCodeExpanded();
      expect(useProjectStore.getState().codeExpanded).toBe(false);
    });
  });

  describe('Session Management', () => {
    it('should set session id', () => {
      useProjectStore.getState().setSessionId('test-session-123');
      expect(useProjectStore.getState().sessionId).toBe('test-session-123');
    });
  });

  describe('Image Management', () => {
    const mockImage: ImageFile = {
      id: 'img-1',
      file: new File([''], 'test.png', { type: 'image/png' }),
      preview: 'blob:test',
      order: 0,
    };

    it('should add single image', () => {
      useProjectStore.getState().addImage(mockImage);
      expect(useProjectStore.getState().images).toHaveLength(1);
      expect(useProjectStore.getState().images[0].id).toBe('img-1');
    });

    it('should add multiple images', () => {
      const images: ImageFile[] = [
        mockImage,
        { ...mockImage, id: 'img-2', order: 1 },
        { ...mockImage, id: 'img-3', order: 2 },
      ];
      useProjectStore.getState().addImages(images);
      expect(useProjectStore.getState().images).toHaveLength(3);
    });

    it('should remove image by id', () => {
      useProjectStore.getState().addImages([
        mockImage,
        { ...mockImage, id: 'img-2', order: 1 },
      ]);
      useProjectStore.getState().removeImage('img-1');
      expect(useProjectStore.getState().images).toHaveLength(1);
      expect(useProjectStore.getState().images[0].id).toBe('img-2');
    });

    it('should reorder images', () => {
      const images: ImageFile[] = [
        { ...mockImage, id: 'img-1', order: 0 },
        { ...mockImage, id: 'img-2', order: 1 },
      ];
      useProjectStore.getState().addImages(images);

      const reordered = [
        { ...mockImage, id: 'img-2', order: 0 },
        { ...mockImage, id: 'img-1', order: 1 },
      ];
      useProjectStore.getState().reorderImages(reordered);

      expect(useProjectStore.getState().images[0].id).toBe('img-2');
      expect(useProjectStore.getState().images[1].id).toBe('img-1');
    });

    it('should clear all images', () => {
      useProjectStore.getState().addImages([mockImage, { ...mockImage, id: 'img-2' }]);
      useProjectStore.getState().clearImages();
      expect(useProjectStore.getState().images).toHaveLength(0);
    });
  });

  describe('Generation State', () => {
    it('should set status', () => {
      useProjectStore.getState().setStatus('generating');
      expect(useProjectStore.getState().status).toBe('generating');
    });

    it('should set thinking content', () => {
      useProjectStore.getState().setThinking('Analyzing images...');
      expect(useProjectStore.getState().thinkingContent).toBe('Analyzing images...');
    });

    it('should append thinking content', () => {
      useProjectStore.getState().setThinking('Part 1. ');
      useProjectStore.getState().appendThinking('Part 2.');
      expect(useProjectStore.getState().thinkingContent).toBe('Part 1. Part 2.');
    });

    it('should set generated code and update status to completed', () => {
      const code = { code: 'const App = () => <div>Hello</div>;', timestamp: Date.now() };
      useProjectStore.getState().setGeneratedCode(code);
      expect(useProjectStore.getState().generatedCode?.code).toBe(code.code);
      expect(useProjectStore.getState().status).toBe('completed');
    });

    it('should update code', () => {
      const initialCode = { code: 'initial', timestamp: Date.now() };
      useProjectStore.getState().setGeneratedCode(initialCode);
      useProjectStore.getState().updateCode('updated code');
      expect(useProjectStore.getState().generatedCode?.code).toBe('updated code');
    });

    it('should update code when no existing code', () => {
      useProjectStore.getState().updateCode('new code');
      expect(useProjectStore.getState().generatedCode?.code).toBe('new code');
    });

    it('should update CSS', () => {
      useProjectStore.getState().updateCss('.app { color: red; }');
      expect(useProjectStore.getState().generatedCss).toBe('.app { color: red; }');
    });

    it('should set error', () => {
      useProjectStore.getState().setError('Something went wrong');
      expect(useProjectStore.getState().status).toBe('error');
      expect(useProjectStore.getState().errorMessage).toBe('Something went wrong');
    });

    it('should reset generation state', () => {
      useProjectStore.getState().setGeneratedCode({ code: 'test', timestamp: Date.now() });
      useProjectStore.getState().setThinking('thinking...');
      useProjectStore.getState().updateCss('css');

      useProjectStore.getState().reset();

      expect(useProjectStore.getState().generatedCode).toBeNull();
      expect(useProjectStore.getState().thinkingContent).toBe('');
      expect(useProjectStore.getState().generatedCss).toBe('');
      expect(useProjectStore.getState().status).toBe('idle');
    });
  });

  describe('Image IDs', () => {
    it('should set image ids', () => {
      useProjectStore.getState().setImageIds(['id1', 'id2', 'id3']);
      expect(useProjectStore.getState().imageIds).toEqual(['id1', 'id2', 'id3']);
    });
  });

  describe('Editor State', () => {
    it('should set active file', () => {
      useProjectStore.getState().setActiveFile('styles.css');
      expect(useProjectStore.getState().activeFile).toBe('styles.css');
    });
  });

  describe('Chat Messages', () => {
    it('should add user message', () => {
      useProjectStore.getState().addUserMessage('Hello');
      const messages = useProjectStore.getState().chatMessages;
      expect(messages).toHaveLength(1);
      expect(messages[0].role).toBe('user');
      expect(messages[0].content).toBe('Hello');
    });

    it('should add user message with images', () => {
      useProjectStore.getState().addUserMessage('Check this', ['img1.png', 'img2.png']);
      const messages = useProjectStore.getState().chatMessages;
      expect(messages).toHaveLength(1);
      expect(messages[0].images).toEqual(['img1.png', 'img2.png']);
    });

    it('should add assistant message', () => {
      useProjectStore.getState().addAssistantMessage('Hello, how can I help?');
      const messages = useProjectStore.getState().chatMessages;
      expect(messages).toHaveLength(1);
      expect(messages[0].role).toBe('assistant');
      expect(messages[0].content).toBe('Hello, how can I help?');
    });

    it('should update last assistant message', () => {
      useProjectStore.getState().addAssistantMessage('Initial');
      useProjectStore.getState().updateLastAssistantMessage('Updated');
      const messages = useProjectStore.getState().chatMessages;
      expect(messages).toHaveLength(1);
      expect(messages[0].content).toBe('Updated');
    });

    it('should append to last assistant message', () => {
      useProjectStore.getState().addAssistantMessage('Part 1');
      useProjectStore.getState().appendToLastAssistantMessage(' Part 2');
      const messages = useProjectStore.getState().chatMessages;
      expect(messages).toHaveLength(1);
      expect(messages[0].content).toBe('Part 1 Part 2');
    });

    it('should clear all chat messages', () => {
      useProjectStore.getState().addUserMessage('Hello');
      useProjectStore.getState().addAssistantMessage('Hi');
      useProjectStore.getState().clearChatMessages();
      expect(useProjectStore.getState().chatMessages).toHaveLength(0);
    });

    it('should not update if no assistant message exists', () => {
      useProjectStore.getState().addUserMessage('Hello');
      useProjectStore.getState().updateLastAssistantMessage('Updated');
      const messages = useProjectStore.getState().chatMessages;
      expect(messages).toHaveLength(1);
      expect(messages[0].content).toBe('Hello');
    });

    it('should not append if no assistant message exists', () => {
      useProjectStore.getState().addUserMessage('Hello');
      useProjectStore.getState().appendToLastAssistantMessage(' more');
      const messages = useProjectStore.getState().chatMessages;
      expect(messages).toHaveLength(1);
      expect(messages[0].content).toBe('Hello');
    });
  });
});
