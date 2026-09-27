"use client";

import { LoaderCircleIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { useDeleteCategory } from "@/features/categories/hooks/use-delete-category";
import type { Category } from "@/features/categories/types/category";

export function DeleteCategoryDialog({ category }: { category: Category }) {
  const [open, setOpen] = useState(false);
  const deleteCategory = useDeleteCategory();

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      deleteCategory.reset();
    }
  }

  async function handleDelete() {
    try {
      await deleteCategory.mutateAsync(category.id);
      handleOpenChange(false);
    } catch {
      // The mutation error is rendered in the confirmation dialog.
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          aria-label={`Delete ${category.name}`}
          className="text-muted-foreground hover:text-destructive"
        >
          <Trash2Icon aria-hidden="true" />
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete {category.name}?</AlertDialogTitle>
          <AlertDialogDescription>
            This permanently removes the custom category. This action cannot be
            undone.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {deleteCategory.isError ? (
          <p
            role="alert"
            className="mt-4 rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
          >
            We could not delete this category. It may be in use by a transaction.
          </p>
        ) : null}

        <AlertDialogFooter>
          <AlertDialogCancel asChild>
            <Button variant="outline" disabled={deleteCategory.isPending}>
              Cancel
            </Button>
          </AlertDialogCancel>
          <AlertDialogAction asChild>
            <Button
              variant="destructive"
              disabled={deleteCategory.isPending}
              onClick={(event) => {
                event.preventDefault();
                void handleDelete();
              }}
            >
              {deleteCategory.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {deleteCategory.isPending ? "Deleting…" : "Delete category"}
            </Button>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
