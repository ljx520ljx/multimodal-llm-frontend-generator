'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { useProjectStore } from '@/stores/useProjectStore';
import { useDebouncedValue } from '@/lib/hooks';

// 元素信息类型
export interface SelectedElementInfo {
  selector: string;       // CSS 选择器
  tagName: string;        // 标签名
  text: string;           // 文本内容
  classList: string[];    // class 列表
  styles: Record<string, string>;  // 关键样式
}

const DEFAULT_HTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>原型预览</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="text-center">
      <h1 class="text-2xl font-bold text-gray-800 mb-4">
        上传设计稿开始生成
      </h1>
      <p class="text-gray-600">
        支持交互预览，点击体验生成的原型
      </p>
    </div>
  </div>
</body>
</html>`;

interface HtmlPreviewProps {
  onRefresh?: () => void;
  isGenerating?: boolean;
  annotationMode?: boolean;  // 标注模式
  onElementSelect?: (info: SelectedElementInfo) => void;  // 元素选中回调
}

// 注入到 iframe 的标注模式脚本
const ANNOTATION_SCRIPT = `
<script>
(function() {
  let annotationMode = false;
  let hoveredElement = null;

  // 阻止 JS 导航（window.open, window.location 赋值, assign, replace 等）
  window.open = function() { return null; };
  try {
    Object.defineProperty(window, 'location', {
      configurable: false,
      get: function() { return location; },
      set: function() { /* 阻止 window.location = ... */ }
    });
  } catch(e) { /* 某些环境不允许重定义 */ }
  var origAssign = Location.prototype.assign;
  var origReplace = Location.prototype.replace;
  Location.prototype.assign = function() { /* blocked */ };
  Location.prototype.replace = function() { /* blocked */ };

  // 生成元素的 CSS 选择器
  function getSelector(el) {
    if (el.id) return '#' + el.id;

    let path = [];
    while (el && el.nodeType === Node.ELEMENT_NODE) {
      let selector = el.tagName.toLowerCase();
      if (el.className && typeof el.className === 'string') {
        const classes = el.className.trim().split(/\\s+/).filter(c => c && !c.startsWith('__'));
        if (classes.length > 0) {
          selector += '.' + classes.slice(0, 2).join('.');
        }
      }
      path.unshift(selector);
      if (el.parentElement === document.body || path.length > 3) break;
      el = el.parentElement;
    }
    return path.join(' > ');
  }

  // 获取元素关键样式
  function getKeyStyles(el) {
    const computed = window.getComputedStyle(el);
    return {
      color: computed.color,
      backgroundColor: computed.backgroundColor,
      fontSize: computed.fontSize,
      fontWeight: computed.fontWeight,
      padding: computed.padding,
      margin: computed.margin,
      borderRadius: computed.borderRadius,
    };
  }

  // 高亮元素
  function highlightElement(el) {
    if (hoveredElement) {
      hoveredElement.style.outline = '';
      hoveredElement.style.outlineOffset = '';
    }
    if (el && el !== document.body && el !== document.documentElement) {
      el.style.outline = '2px solid #3b82f6';
      el.style.outlineOffset = '2px';
      hoveredElement = el;
    }
  }

  // 监听来自父页面的消息
  window.addEventListener('message', function(e) {
    // srcdoc iframe 的 origin 为 'null'，无法做可靠的 origin 校验
    // 安全由 iframe sandbox 属性 + 消息类型白名单保证
    if (!e.data || typeof e.data !== 'object') return;
    if (e.data.type === 'setAnnotationMode') {
      annotationMode = e.data.enabled;
      if (!annotationMode && hoveredElement) {
        hoveredElement.style.outline = '';
        hoveredElement.style.outlineOffset = '';
        hoveredElement = null;
      }
      document.body.style.cursor = annotationMode ? 'crosshair' : '';
    }
  });

  // 鼠标移动高亮
  document.addEventListener('mousemove', function(e) {
    if (!annotationMode) return;
    highlightElement(e.target);
  });

  // 拦截所有可能导致导航的点击，防止 iframe 内导航
  document.addEventListener('click', function(e) {
    const link = e.target.closest('a[href]');
    if (link) {
      const href = link.getAttribute('href');
      // 阻止所有链接导航（外部链接、绝对路径、锚点、空链接等）
      // 只放行 Alpine.js 内部状态切换（无 href 的按钮、@click 处理的元素）
      if (href !== null && href !== undefined) {
        e.preventDefault();
      }
    }

    // 拦截所有表单提交导航（不仅限于 http/slash action）
    const form = e.target.closest('form');
    if (form && (e.target.type === 'submit' || e.target.closest('[type="submit"]'))) {
      e.preventDefault();
    }

    // 标注模式：选中元素
    if (!annotationMode) return;
    e.preventDefault();
    e.stopPropagation();

    const el = e.target;
    if (el === document.body || el === document.documentElement) return;

    const info = {
      selector: getSelector(el),
      tagName: el.tagName.toLowerCase(),
      text: el.innerText?.slice(0, 100) || '',
      classList: Array.from(el.classList || []),
      styles: getKeyStyles(el),
    };

    window.parent.postMessage({ type: 'elementSelected', info: info }, '*');
  }, true);

  // 拦截所有 form submit 事件
  document.addEventListener('submit', function(e) {
    e.preventDefault();
  }, true);

  // hashchange 安全网：阻止 hash 变化触发的导航行为
  window.addEventListener('hashchange', function(e) {
    e.preventDefault();
  });
})();
</script>
`;

export function HtmlPreview({ onRefresh, isGenerating: _isGenerating = false, annotationMode = false, onElementSelect }: HtmlPreviewProps) {
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const rawCode = generatedCode?.code || DEFAULT_HTML;

  // 使用防抖值，避免频繁重渲染
  const debouncedCode = useDebouncedValue(rawCode, 300);

  // 验证 HTML 代码
  const validateHtml = useCallback((html: string): boolean => {
    // 基本验证：检查是否包含必要的 HTML 结构或有效的 HTML 标签
    const hasStructure = html.includes('<html') || html.includes('<body') || html.includes('<!DOCTYPE');
    const hasValidTags = /<\w+[^>]*>/.test(html) && html.length > 50;
    return hasStructure || hasValidTags;
  }, []);

  // 更新 iframe 内容
  useEffect(() => {
    if (!iframeRef.current) return;

    setIsLoading(true);
    setError(null);

    try {
      // 验证代码
      if (!validateHtml(debouncedCode)) {
        setError('代码格式无效');
        setIsLoading(false);
        return;
      }

      // 确保代码是完整的 HTML
      let finalHtml = debouncedCode;

      // 如果代码不包含完整的 HTML 结构，包装它
      if (!finalHtml.includes('<!DOCTYPE') && !finalHtml.includes('<html')) {
        finalHtml = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
${finalHtml}
</body>
</html>`;
      }

      // 注入标注模式脚本到 <head> 内部（紧跟 <head...> 之后）
      // 原来注入在 </body> 前，但 LLM 生成的 HTML 可能有未闭合的
      // <style>/<script>/<textarea> 等元素，导致浏览器将注入的脚本
      // 当作这些未闭合元素的文本内容解析，脚本不执行反而显示为可见文本。
      // 注入到 <head> 内可避免此问题，因为 head 在 body 之前解析。
      const headMatch = finalHtml.match(/<head[^>]*>/i);
      if (headMatch) {
        const insertPos = finalHtml.indexOf(headMatch[0]) + headMatch[0].length;
        finalHtml = finalHtml.substring(0, insertPos) + ANNOTATION_SCRIPT + finalHtml.substring(insertPos);
      } else {
        // fallback: 尝试在 <html...> 后插入
        const htmlMatch = finalHtml.match(/<html[^>]*>/i);
        if (htmlMatch) {
          const insertPos = finalHtml.indexOf(htmlMatch[0]) + htmlMatch[0].length;
          finalHtml = finalHtml.substring(0, insertPos) + ANNOTATION_SCRIPT + finalHtml.substring(insertPos);
        } else {
          // 最终 fallback: prepend
          finalHtml = ANNOTATION_SCRIPT + finalHtml;
        }
      }

      // 使用 srcdoc 设置 iframe 内容
      iframeRef.current.srcdoc = finalHtml;
    } catch (err) {
      setError(err instanceof Error ? err.message : '渲染失败');
    }
  }, [debouncedCode, validateHtml]);

  // 监听 iframe 加载完成
  const handleLoad = useCallback(() => {
    setIsLoading(false);
    // 加载完成后设置标注模式状态
    if (iframeRef.current?.contentWindow) {
      iframeRef.current.contentWindow.postMessage({ type: 'setAnnotationMode', enabled: annotationMode }, '*');
    }
  }, [annotationMode]);

  // 当标注模式改变时通知 iframe
  useEffect(() => {
    if (iframeRef.current?.contentWindow) {
      iframeRef.current.contentWindow.postMessage({ type: 'setAnnotationMode', enabled: annotationMode }, '*');
    }
  }, [annotationMode]);

  // 监听 iframe 发来的元素选中消息
  useEffect(() => {
    const handleMessage = (e: MessageEvent) => {
      // 验证消息来源是我们的 iframe
      if (e.source !== iframeRef.current?.contentWindow) return;
      if (e.data.type === 'elementSelected' && onElementSelect) {
        onElementSelect(e.data.info);
      }
    };
    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, [onElementSelect]);

  return (
    <div className="relative h-full w-full flex flex-col" style={{ minHeight: '200px' }}>
      {/* 加载状态 */}
      {isLoading && (
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-white/90 z-10">
          <svg className="h-8 w-8 animate-spin text-blue-500 mb-3" viewBox="0 0 24 24">
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
              fill="none"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
            />
          </svg>
          <p className="text-sm text-slate-600">正在加载预览...</p>
        </div>
      )}

      {/* 错误状态 */}
      {error && (
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-red-50 z-10">
          <svg
            className="mb-3 h-10 w-10 text-red-400"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <circle cx="12" cy="12" r="10" />
            <path d="M12 8v4M12 16h.01" />
          </svg>
          <p className="text-sm font-medium text-red-700">渲染错误</p>
          <p className="mt-2 max-w-md text-center text-xs text-red-600">{error}</p>
          {onRefresh && (
            <button
              onClick={onRefresh}
              className="mt-3 rounded bg-red-100 px-3 py-1 text-xs text-red-700 hover:bg-red-200"
            >
              重新加载
            </button>
          )}
        </div>
      )}

      {/* iframe 预览 */}
      <iframe
        ref={iframeRef}
        className="flex-1 w-full border-0"
        sandbox="allow-scripts"
        onLoad={handleLoad}
        title="原型预览"
      />
    </div>
  );
}
