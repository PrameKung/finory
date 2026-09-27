"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { LoaderCircleIcon, PlusIcon } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { createCategorySchema } from "@/features/categories/schemas/category-schema";
import { useCreateCategory } from "@/features/categories/hooks/use-create-category";
import type { CreateCategoryInput } from "@/features/categories/types/category";
import { ApiError } from "@/lib/api/client";
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

function mutationErrorMessage(error: Error | null) {
  if (error instanceof ApiError && error.status === 409) {
    return "A category with this name and type already exists.";
  }

  if (error instanceof ApiError && error.status === 400) {
    return "Check the category details and try again.";
  }

  return "We could not create this category. Please try again.";
}

export function CreateCategorySheet() {
  const [open, setOpen] = useState(false);
  const createCategory = useCreateCategory();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateCategoryInput>({
    resolver: zodResolver(createCategorySchema),
    defaultValues: {
      name: "",
      type: "expense",
    },
  });

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      reset();
      createCategory.reset();
    }
  }

  async function onSubmit(input: CreateCategoryInput) {
    try {
      await createCategory.mutateAsync(input);
      handleOpenChange(false);
    } catch {
      // The mutation error is rendered below the form fields.
    }
  }

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetTrigger asChild>
        <Button type="button" size="lg">
          <PlusIcon data-icon="inline-start" aria-hidden="true" />
          New category
        </Button>
      </SheetTrigger>
      <SheetContent className="w-[min(26rem,100vw)] sm:max-w-md">
        <SheetHeader className="border-b px-5 py-5 pr-14">
          <SheetTitle className="text-lg">Create category</SheetTitle>
          <SheetDescription>
            Add a custom category for organizing your transactions.
          </SheetDescription>
        </SheetHeader>

        <form
          id="create-category-form"
          className="flex flex-1 flex-col"
          onSubmit={handleSubmit(onSubmit)}
          noValidate
        >
          <div className="flex-1 space-y-6 overflow-y-auto px-5 py-6">
            <div className="space-y-2">
              <label htmlFor="category-name" className="text-sm font-medium">
                Name
              </label>
              <input
                id="category-name"
                type="text"
                autoComplete="off"
                maxLength={100}
                placeholder="e.g. Education"
                aria-invalid={Boolean(errors.name)}
                aria-describedby={errors.name ? "category-name-error" : undefined}
                className="h-10 w-full rounded-lg border bg-background px-3 text-sm outline-none transition-shadow placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20"
                {...register("name")}
              />
              {errors.name ? (
                <p id="category-name-error" className="text-sm text-destructive">
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

            {createCategory.isError ? (
              <p
                role="alert"
                className="rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
              >
                {mutationErrorMessage(createCategory.error)}
              </p>
            ) : null}
          </div>

          <SheetFooter className="border-t px-5 py-4 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={createCategory.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createCategory.isPending}>
              {createCategory.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {createCategory.isPending ? "Creating…" : "Create category"}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
