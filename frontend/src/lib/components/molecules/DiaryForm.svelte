<script lang="ts">
import { goto } from "$app/navigation";

export let title: string;
export let content = "";
export let date = "";
export let error: string | undefined = undefined;
export let showDeleteButton = false;
export let onCancel: (() => void) | null = null;
export let onDelete: (() => void) | null = null;

function _handleCancel() {
	if (onCancel) {
		onCancel();
	} else {
		goto("/");
	}
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">{title}</h1>
		<Button
			variant="secondary"
			size="md"
			on:click={_handleCancel}
		>
			{$_('diary.back')}
		</Button>
	</div>

	<div class="bg-white shadow rounded-lg p-6">
		<form method="POST" use:enhance>
			{#if error}
				<Alert type="error">
					{error}
				</Alert>
			{/if}

			<FormField
				type="input"
				inputType="date"
				label={$_('create.date')}
				id="date"
				name="date"
				required
				bind:value={date}
			/>

			<FormField
				type="textarea"
				label={$_('create.content')}
				id="content"
				name="content"
				placeholder={$_('diary.placeholder')}
				required
				rows={12}
				bind:value={content}
			/>

			<div class="flex {showDeleteButton ? 'justify-between' : 'justify-end'}">
				{#if showDeleteButton && onDelete}
					<Button
						type="button"
						variant="danger"
						size="md"
						on:click={onDelete}
					>
						{$_('diary.delete')}
					</Button>
				{/if}

				<div class="flex space-x-4">
					<Button
						type="button"
						variant="secondary"
						size="md"
						on:click={_handleCancel}
					>
						{$_('diary.cancel')}
					</Button>
					<Button
						type="submit"
						variant="primary"
						size="md"
					>
						{$_('diary.save')}
					</Button>
				</div>
			</div>
		</form>
	</div>
</div>