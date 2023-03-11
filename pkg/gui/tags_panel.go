package gui

import "github.com/jesseduffield/lazygit/pkg/gui/types"

func (gui *Gui) tagsRenderToMain() error {
	pair := gui.c.MainViewPairs().Normal
	width := pair.Main.GetView().Width() - 40
	var task types.UpdateTask
	tag := gui.State.Contexts.Tags.GetSelected()
	if tag == nil {
		task = types.NewRenderStringTask("No tags")
	} else {
		cmdObj := gui.git.Branch.GetGraphCmdObj(tag.FullRefName(), width)
		task = types.NewRunCommandTask(cmdObj.GetCmd())
	}

	return gui.c.RenderToMainViews(types.RefreshMainOpts{
		Pair: pair,
		Main: &types.ViewUpdateOpts{
			Title: "Tag",
			Task:  task,
		},
	})
}
