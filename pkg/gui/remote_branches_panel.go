package gui

import "github.com/jesseduffield/lazygit/pkg/gui/types"

func (gui *Gui) remoteBranchesRenderToMain() error {
	pair := gui.c.MainViewPairs().Normal
	width := pair.Main.GetView().Width() - 40
	var task types.UpdateTask
	remoteBranch := gui.State.Contexts.RemoteBranches.GetSelected()
	if remoteBranch == nil {
		task = types.NewRenderStringTask("No branches for this remote")
	} else {
		cmdObj := gui.git.Branch.GetGraphCmdObj(remoteBranch.FullRefName(), width)
		task = types.NewRunCommandTask(cmdObj.GetCmd())
	}

	return gui.c.RenderToMainViews(types.RefreshMainOpts{
		Pair: pair,
		Main: &types.ViewUpdateOpts{
			Title: "Remote Branch",
			Task:  task,
		},
	})
}
