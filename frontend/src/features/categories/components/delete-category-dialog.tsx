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
import { InlineMessage } from "@/components/shared/inline-message";
import { useDeleteCategory } from "@/features/categories/hooks/use-delete-category";
import type { Category } from "@/features/categories/types/category";
import { getApiErrorMessage } from "@/lib/api/client";

function mutationErrorMessage(error: unknown) {
  return getApiErrorMessage(error, {
    defaultMessage:
      "We could not delete this category. It may be in use by a transaction.",
    codeMessages: {
      category_not_found:
        "This category no longer exists or has already been deleted.",
    },
  });
}

export function DeleteCategoryDialog({ category }: { category: Category }) {
  const [open, setOpen] = useState(false);
  const deleteCategory = useDeleteCategory();

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen && deleteCategory.isPending) {
      return;
    }

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
          <InlineMessage className="mt-4">
            {mutationErrorMessage(deleteCategory.error)}
          </InlineMessage>
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
