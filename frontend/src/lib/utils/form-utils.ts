import { invalidateAll } from "$app/navigation";

export function createSubmitHandler(
	setLoading: (loading: boolean) => void,
	setSaved?: (saved: boolean) => void
) {
	return () => {
		setLoading(true);
		return async ({ result }: any) => {
			setLoading(false);
			if (result.type === "success") {
				await invalidateAll();
				if (setSaved) {
					setSaved(true);
					setTimeout(() => setSaved(false), 1000);
				}
			}
		};
	};
}