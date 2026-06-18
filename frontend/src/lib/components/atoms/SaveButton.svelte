<script lang="ts">
  import { _ } from "svelte-i18n";
  import { onMount, onDestroy } from "svelte";
  import Button from "./Button.svelte";
  import { attachHapticToButton } from "$lib/utils/haptic";

  export let loading = false;
  export let saved = false;
  export let size: "sm" | "md" | "lg" = "md";
  // カスタムラベル（未指定時はデフォルトのi18nキーを使用）
  export let label: string | null = null;
  // type="submit"か"button"かを選択可能にする
  export let type: "submit" | "button" = "submit";

  let buttonEl: HTMLElement;
  let detach: (() => void) | null = null;
  let wrapperEl: HTMLElement;

  // パーティクルの型定義
  type Particle = {
    id: number;
    x: number;
    y: number;
    vx: number;
    vy: number;
    size: number;
    opacity: number;
  };

  let particles: Particle[] = [];
  let particleId = 0;
  let animationFrame: ReturnType<typeof requestAnimationFrame> | null = null;

  // saved が true になったタイミングでパーティクルを生成する
  let prevSaved = false;
  $: {
    if (saved && !prevSaved) {
      spawnParticles();
    }
    prevSaved = saved;
  }

  function spawnParticles() {
    if (!wrapperEl) return;

    const rect = wrapperEl.getBoundingClientRect();
    // ボタン中央を原点にする（ラッパー相対座標）
    const cx = rect.width / 2;
    const cy = rect.height / 2;

    const count = 27;
    // ボタンを楕円と見立てて、楕円の周囲を均等分割した位置から外側へ飛び出す
    const rx = rect.width / 2;
    const ry = rect.height / 2;
    const newParticles: Particle[] = Array.from({ length: count }, (_, i) => {
      const angle = (i / count) * Math.PI * 2 + (Math.random() - 0.5) * 0.3;
      // 楕円上の発生位置
      const px = cx + rx * Math.cos(angle);
      const py = cy + ry * Math.sin(angle);
      // 拡散範囲を0.7倍に抑えた速度
      const speed = 1.5 + Math.random() * 2;
      return {
        id: particleId++,
        x: px,
        y: py,
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed,
        size: 1.5 + Math.random(),
        opacity: 1,
      };
    });

    particles = [...particles, ...newParticles];

    if (animationFrame === null) {
      animationFrame = requestAnimationFrame(animateParticles);
    }
  }

  function animateParticles() {
    // 重力なし・減衰のみで全方向に広がる
    particles = particles
      .map((p) => ({
        ...p,
        x: p.x + p.vx,
        y: p.y + p.vy,
        vx: p.vx * 0.94,
        vy: p.vy * 0.94,
        opacity: p.opacity - 0.035,
      }))
      .filter((p) => p.opacity > 0);

    if (particles.length > 0) {
      animationFrame = requestAnimationFrame(animateParticles);
    } else {
      animationFrame = null;
    }
  }

  onMount(() => {
    // ボタン押下（pointerdown）のタイミングで振動させる
    // iOS Safari は非同期コールバック内では haptic が無効になる場合があるため、
    // ユーザージェスチャーが確実に存在する pointerdown で発火する
    detach = attachHapticToButton(buttonEl);
  });

  onDestroy(() => {
    detach?.();
    if (animationFrame !== null) {
      cancelAnimationFrame(animationFrame);
    }
  });
</script>

<!--
  on:clickイベントをバブルアップ
  親コンポーネントがtype="button"で独自のクリックハンドラを指定できるようにする
-->
<div class="relative inline-block" bind:this={wrapperEl}>
  <!-- パーティクルを描画するSVGレイヤー（ボタンの上にpointer-events:noneで重ねる） -->
  {#if particles.length > 0}
    <svg
      class="absolute inset-0 w-full h-full pointer-events-none overflow-visible"
      style="z-index: 50;"
    >
      {#each particles as p (p.id)}
        <circle
          cx={p.x}
          cy={p.y}
          r={p.size}
          fill="white"
          opacity={p.opacity}
        />
      {/each}
    </svg>
  {/if}

  <div bind:this={buttonEl}>
    <Button {type} variant={saved ? "success" : "primary"} {size} disabled={loading || saved} on:click>
      <div class="flex items-center justify-center min-h-[1.25rem] min-w-[4.5rem]">
        {#if loading}
          <svg class="animate-spin -mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 714 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
          </svg>
          <span class="ml-1">{$_("diary.saving")}</span>
        {:else if saved}
          <svg class="-mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
            <path stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" d="m9 12 2 2 4-4"/>
          </svg>
          <span class="ml-1">{$_("diary.saved")}</span>
        {:else}
          <span>{label !== null ? label : $_("diary.save")}</span>
        {/if}
      </div>
    </Button>
  </div>
</div>
