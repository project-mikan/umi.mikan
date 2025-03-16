<script lang="ts">
  import Logo from "$lib/assets/logo_v1.png";
  import NormalButton from "$lib/components/atom/button/NormalButton.svelte";
</script>

<h1 class="text-2xl m-4 text-center">umi.mikan</h1>

<div class="flex justify-center">
  <img alt="logo" src={Logo} loading="lazy" class="w-24 h-24" />
</div>
<NormalButton
  title="login"
  url="/login"
  bgColorVariable="blue"
  textColorVariable="black"
/>
