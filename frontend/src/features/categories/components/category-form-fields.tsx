import type { FieldErrors, UseFormRegister } from "react-hook-form";

import type { CreateCategoryInput } from "@/features/categories/types/category";
import { cn } from "@/lib/utils";

const categoryTypes = [
  {
    value: "expense",
    label: "Expense",
    description: "Money going out",
  },
  {
    value: "income",
    label: "Income",
    description: "Money coming in",
  },
] as const;

type CategoryFormFieldsProps = {
  nameId: string;
  register: UseFormRegister<CreateCategoryInput>;
  errors: FieldErrors<CreateCategoryInput>;
};

export function CategoryFormFields({
  nameId,
  register,
  errors,
}: CategoryFormFieldsProps) {
  const errorId = `${nameId}-error`;

  return (
    <>
      <div className="space-y-2">
        <label htmlFor={nameId} className="text-sm font-medium">
          Name
        </label>
        <input
          id={nameId}
          type="text"
          autoComplete="off"
          maxLength={100}
          placeholder="e.g. Education"
          aria-invalid={Boolean(errors.name)}
          aria-describedby={errors.name ? errorId : undefined}
          className="h-10 w-full rounded-lg border bg-background px-3 text-sm outline-none transition-shadow placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20"
          {...register("name")}
        />
        {errors.name ? (
          <p id={errorId} className="text-sm text-destructive">
            {errors.name.message}
          </p>
        ) : null}
      </div>

      <fieldset className="space-y-2">
        <legend className="text-sm font-medium">Type</legend>
        <div className="grid grid-cols-2 gap-3">
          {categoryTypes.map((type) => (
            <label
              key={type.value}
              className={cn(
                "has-checked:border-foreground has-checked:bg-muted/60 flex cursor-pointer flex-col rounded-xl border p-3 transition-colors hover:bg-muted/40",
                "has-focus-visible:ring-3 has-focus-visible:ring-ring/50",
              )}
            >
              <input
                type="radio"
                value={type.value}
                className="sr-only"
                {...register("type")}
              />
              <span className="text-sm font-medium">{type.label}</span>
              <span className="text-xs text-muted-foreground">
                {type.description}
              </span>
            </label>
          ))}
        </div>
      </fieldset>
    </>
  );
}
